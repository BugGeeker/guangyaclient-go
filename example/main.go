package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	guangyaclient "github.com/BugGeeker/guangyaclient-go"
)

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
	UpdatedAt    string `json:"updated_at"`
}

type record struct {
	Name      string `json:"name"`
	StartedAt string `json:"started_at"`
	Duration  string `json:"duration"`
	Success   bool   `json:"success"`
	Result    any    `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
}

type runLog struct {
	StartedAt  string   `json:"started_at"`
	FinishedAt string   `json:"finished_at"`
	Records    []record `json:"records"`
}

func main() {
	phone := flag.String("phone", os.Getenv("GUANGYA_PHONE"), "短信登录手机号")
	fileID := flag.String("file-id", os.Getenv("GUANGYA_FILE_ID"), "文件 ID")
	shareID := flag.String("share-id", os.Getenv("GUANGYA_SHARE_ID"), "分享 ID")
	shareIDStr := flag.String("share-id-str", os.Getenv("GUANGYA_SHARE_ID_STR"), "分享 ID 字符串")
	shareCode := flag.String("share-code", os.Getenv("GUANGYA_SHARE_CODE"), "分享提取码")
	cloudURL := flag.String("cloud-url", os.Getenv("GUANGYA_CLOUD_URL"), "云添加地址")
	uploadPath := flag.String("upload", os.Getenv("GUANGYA_UPLOAD"), "上传文件路径")
	torrentPath := flag.String("torrent", os.Getenv("GUANGYA_TORRENT"), "种子文件路径")
	parentId := flag.String("parent-id", os.Getenv("GUANGYA_PARENT_ID"), "父目录 ID")
	runDestructive := flag.Bool("run-destructive", false, "执行创建目录、复制、移动、删除、回收站等变更操作")
	flag.Parse()

	if err := os.MkdirAll("logs", 0o755); err != nil {
		log.Fatal(err)
	}
	saved, err := loadTokens("logs/tokens.json")
	if err != nil {
		log.Fatal(err)
	}
	client := guangyaclient.NewClient(saved.AccessToken, saved.RefreshToken, saved.DeviceID)
	defer client.Close()

	run := &runLog{StartedAt: time.Now().Format(time.RFC3339)}
	if err := authenticate(client, run, *phone); err != nil {
		fmt.Println("认证失败:", err)
		addError(run, "authentication", err)
		writeRunLog(run)
		log.Fatal(err)
	}
	if err := saveTokens("logs/tokens.json", client); err != nil {
		log.Fatal(err)
	}

	// 查询类接口默认全部执行。
	call(run, "user_info", client.UserInfo)
	call(run, "get_assets", client.GetAssets)
	call(run, "fs_files", func() (guangyaclient.FileListResponse, error) {
		return client.FSFiles(nil, 0, 50, 0, 0, nil, nil, nil, false)
	})
	call(run, "fs_image_list", func() (guangyaclient.FileListResponse, error) { return client.FSImageList(0, 50, 3, 1) })
	call(run, "fs_video_list", func() (guangyaclient.FileListResponse, error) { return client.FSVideoList(0, 50, 3, 1, true) })
	call(run, "fs_document_list", func() (guangyaclient.FileListResponse, error) { return client.FSDocumentList(0, 50, 3, 1) })
	call(run, "fs_recycle_files", func() (guangyaclient.FileListResponse, error) { return client.FSRecycleFiles(0, 50, 10, 0) })
	call(run, "cloud_task_list", func() (guangyaclient.CloudTaskListResponse, error) { return client.CloudTaskList(0, 50, nil) })
	call(run, "share_user_list", func() (guangyaclient.ShareListResponse, error) { return client.ShareUserList(0, 50, 1, 1) })

	if *fileID != "" {
		call(run, "fs_detail", func() (guangyaclient.FileDetailResponse, error) { return client.FSDetail(*fileID) })
		call(run, "download_url", func() (guangyaclient.DownloadURLResponse, error) { return client.DownloadURL(*fileID) })
		call(run, "share_create", func() (guangyaclient.ShareCreateResponse, error) {
			return client.ShareCreate([]any{*fileID}, "Go 示例分享")
		})
	}
	if *shareIDStr != "" {
		call(run, "share_summary", func() (guangyaclient.ShareSummaryResponse, error) { return client.ShareSummary(*shareIDStr) })
		call(run, "share_update", func() (guangyaclient.EmptyResponse, error) {
			return client.ShareUpdate(*shareID, "Go 示例分享", 0, 1, *shareCode, true, "0", 0, 1)
		})
		if *shareCode != "" {
			var accessToken string
			result, err := call(run, "share_access_token", func() (guangyaclient.ShareAccessTokenResponse, error) {
				return client.ShareAccessToken(*shareIDStr, *shareCode)
			})
			if err == nil {
				accessToken = result.AccessToken
				if accessToken == "" {
					accessToken = result.Data.AccessToken
				}
			}
			if accessToken != "" {
				// call(run, "share_files_list", func() (guangyaclient.ShareFilesListResponse, error) {
				// 	return client.ShareFilesList(accessToken, "", 1, 50, 0, 0)
				// })
				call(run, "share_files_size", func() (guangyaclient.ShareFilesSizeResponse, error) {
					return client.ShareFilesSize(accessToken, []any{*fileID}, true)
				})
				if *fileID != "" {
					call(run, "share_download_url", func() (guangyaclient.ShareDownloadURLResponse, error) {
						return client.ShareDownloadURL(*fileID, accessToken)
					})
				}
			}
		}
	}

	shareAccessToken := ""
	if *shareIDStr != "" && *shareCode != "" {
		if result, err := client.ShareAccessToken(*shareIDStr, *shareCode); err == nil {
			shareAccessToken = result.AccessToken
			if shareAccessToken == "" {
				shareAccessToken = result.Data.AccessToken
			}
		}
	}
	if *runDestructive {
		call(run, "fs_create_dir", func() (guangyaclient.FSCreateDirResponse, error) {
			return client.FSCreateDir("go-example", parentId, false)
		})
		if *fileID != "" {
			call(run, "fs_rename", func() (guangyaclient.EmptyResponse, error) {
				return client.FSRename(*fileID, "go-example-renamed.mp4")
			})
			call(run, "fs_delete", func() (guangyaclient.FileTaskResponse, error) {
				return client.FSDelete([]any{*fileID})
			})
		}
		if *shareID != "" && shareAccessToken != "" {
			call(run, "share_restore", func() (guangyaclient.FileTaskResponse, error) {
				return client.ShareRestore(shareAccessToken, []any{*fileID}, "")
			})
			call(run, "share_delete", func() (guangyaclient.EmptyResponse, error) {
				return client.ShareDelete([]any{*shareID})
			})
		}
	}
	if *cloudURL != "" {
		call(run, "cloud_resolve_url", func() (guangyaclient.CloudResolveURLResponse, error) { return client.CloudResolveURL(*cloudURL) })
		call(run, "cloud_create_task", func() (guangyaclient.CloudCreateTaskResponse, error) {
			return client.CloudCreateTask([]int{}, *cloudURL, parentId, "go-example")
		})
	}
	if *torrentPath != "" {
		data, err := os.ReadFile(*torrentPath)
		if err != nil {
			addError(run, "cloud_resolve_torrent_read", err)
		} else {
			call(run, "cloud_resolve_torrent", func() (guangyaclient.CloudResolveTorrentResponse, error) {
				return client.ResolveTorrent(data, filepath.Base(*torrentPath))
			})
		}
	}
	if *uploadPath != "" {
		call(run, "file_upload", func() (guangyaclient.UploadInfoResponse, error) {
			return client.FileUpload(*uploadPath, "", parentId)
		})
	}

	if *runDestructive {
		call(run, "share_delete", func() (guangyaclient.EmptyResponse, error) {
			return client.ShareDelete([]any{*shareID})
		})
		call(run, "fs_clear_recycle_bin", client.FSClearRecycleBin)
	} else {
		fmt.Println("破坏性接口默认未执行；确认目标后使用 -run-destructive 执行。")
	}
	run.FinishedAt = time.Now().Format(time.RFC3339)
	writeRunLog(run)
}

func authenticate(client *guangyaclient.Client, run *runLog, phone string) error {
	if client.Token != "" {
		if _, err := call(run, "validate_access_token", client.UserInfo); err == nil {
			fmt.Println("access-token 验证成功")
			return nil
		}
	}
	if client.RefreshTokenValue != "" {
		if _, err := call(run, "refresh_token", func() (guangyaclient.TokenResponse, error) {
			return client.RefreshToken("")
		}); err == nil {
			fmt.Println("refresh-token 刷新成功")
			return nil
		}
	}
	if phone == "" {
		return errors.New("未提供有效 token，也未设置 -phone 或 GUANGYA_PHONE")
	}
	_, err := call(run, "login_sms", func() (guangyaclient.TokenResponse, error) {
		return client.LoginSMS(phone, func() (string, error) {
			fmt.Print("请输入短信验证码: ")
			code, err := bufio.NewReader(os.Stdin).ReadString('\n')
			return strings.TrimSpace(code), err
		}, "ANY")
	})
	return err
}

func loadTokens(path string) (tokens, error) {
	var value tokens
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	return value, json.Unmarshal(data, &value)
}

func saveTokens(path string, client *guangyaclient.Client) error {
	data, err := json.MarshalIndent(tokens{
		AccessToken: client.Token, RefreshToken: client.RefreshTokenValue,
		DeviceID: client.DeviceID, UpdatedAt: time.Now().Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func call[T any](run *runLog, name string, fn func() (T, error)) (T, error) {
	start := time.Now()
	result, err := fn()
	item := record{
		Name: name, StartedAt: start.Format(time.RFC3339),
		Duration: time.Since(start).String(), Success: err == nil,
	}
	if err != nil {
		item.Error = err.Error()
		fmt.Printf("[%s] FAIL: %v\n", name, err)
	} else {
		item.Result = result
		fmt.Printf("[%s] OK\n%s\n", name, pretty(result))
	}
	run.Records = append(run.Records, item)
	return result, err
}

func addError(run *runLog, name string, err error) {
	run.Records = append(run.Records, record{
		Name: name, StartedAt: time.Now().Format(time.RFC3339),
		Success: false, Error: err.Error(),
	})
}

func writeRunLog(run *runLog) {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		fmt.Println("JSON 日志序列化失败:", err)
		return
	}
	name := filepath.Join("logs", "run-"+time.Now().Format("20060102-150405")+".json")
	if err := os.WriteFile(name, append(data, '\n'), 0o600); err != nil {
		fmt.Println("JSON 日志写入失败:", err)
		return
	}
	fmt.Println("详细日志:", name)
}

func pretty(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return string(data)
}

var _ = log.Print

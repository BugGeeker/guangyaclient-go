package guangyaclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	apiBase     = "https://api.guangyapan.com"
	accountBase = "https://account.guangyapan.com"
	userAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
)

type Client struct {
	HTTP              *http.Client
	Token             string
	TokenExpiresAt    time.Time
	RefreshTokenValue string
	DeviceID          string
	mu                sync.Mutex
}

func NewClient(accessToken, refreshToken, deviceID string) *Client {
	if deviceID == "" {
		deviceID = GenerateDID()
	}
	return &Client{HTTP: &http.Client{}, Token: accessToken, RefreshTokenValue: refreshToken, DeviceID: deviceID}
}

func (c *Client) Close() {}

// 获取文件列表
// parentId 目录ID
// page 页码
// size 每页数量
// orderBy 排序字段
// sortType 排序方式
// fileTypes 文件类型列表
// resType 资源类型
// dirType 目录类型
// needPlayRecord 是否需要播放记录
func (c *Client) FSFiles(parentId any, page, pageSize, orderBy, sortType int, fileTypes []int, resType, dirType *int, needPlayRecord bool) (FileListResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	payload := FSFilesRequest{
		ParentID: parentId, Page: page, PageSize: pageSize, OrderBy: orderBy, SortType: sortType,
		FileTypes: fileTypes, ResType: resType, DirType: dirType, NeedPlayRecord: needPlayRecord,
	}
	return request[FileListResponse](c, "POST", apiBase+"/userres/v1/file/get_file_list", payload, nil)
}

// 获取云添加任务列表
// page 页码
// pageSize 每页数量
// status 任务状态列表
func (c *Client) CloudTaskList(page, pageSize int, status []int) (CloudTaskListResponse, error) {
	if status == nil {
		status = []int{0, 1, 3, 4}
	}
	return request[CloudTaskListResponse](c, "POST", apiBase+"/nd.bizcloudcollection.s/v1/list_task", CloudTaskListRequest{
		Page: page, PageSize: pageSize, Status: status,
	}, nil)
}

// CloudRetryTask 按状态列表重试云添加任务，不填充默认状态。
func (c *Client) CloudRetryTask(status []int) (CloudRetryTaskResponse, error) {
	return request[CloudRetryTaskResponse](c, "POST", apiBase+"/cloudcollection/v2/retry_task", CloudRetryTaskRequest{
		Status: status,
	}, nil)
}

// CloudDeleteTask 按状态列表删除云添加任务，不填充默认状态。
func (c *Client) CloudDeleteTask(status []int) (CloudDeleteTaskResponse, error) {
	return request[CloudDeleteTaskResponse](c, "POST", apiBase+"/cloudcollection/v2/delete_task", CloudDeleteTaskRequest{
		Status: status,
	}, nil)
}

// 获取上传令牌
// name 文件名
// size 文件大小
// parentId 目录ID
// fileMD5 文件MD5值
func (c *Client) UploadToken(name string, size int64, parentId any, fileMD5 string) (UploadTokenResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	return request[UploadTokenResponse](c, "POST", apiBase+"/userres/v2/get_res_center_token", UploadTokenRequest{
		Capacity: 30, Name: name, Resource: UploadTokenResource{FileSize: size, MD5: fileMD5}, ParentID: parentId,
	}, nil)
}

// 检查是否可以秒传
// taskID 任务ID
// path 文件路径
func (c *Client) CheckCanFlashUpload(taskID, path string) (FlashUploadResponse, error) {
	gcid, err := CalculateGCID(path)
	if err != nil {
		return FlashUploadResponse{}, err
	}
	cid, err := CalculateCID(path)
	if err != nil {
		return FlashUploadResponse{}, err
	}
	return request[FlashUploadResponse](c, "POST", apiBase+"/userres/v1/check_can_flash_upload", CheckCanFlashUploadRequest{
		TaskID: taskID, GCID: gcid, CID: cid,
	}, nil)
}

// 初始化短信登录
// phone 手机号
// captchaToken 验证码token
func (c *Client) LoginSMSInit(phone, captchaToken string) (CaptchaInitResponse, error) {
	body := CaptchaInitRequest{
		ClientID: clientID, Action: "POST:/v1/auth/verification", DeviceID: c.DeviceID,
		Meta: CaptchaMeta{PhoneNumber: phone}, CaptchaToken: captchaToken,
	}
	return request[CaptchaInitResponse](c, "POST", accountBase+"/v1/shield/captcha/init", body, c.accountHeaders())
}

// 发送短信验证码
// phone 手机号
// captchaToken 验证码token
// target 目标
func (c *Client) LoginSMSSend(phone, captchaToken, target string) (SMSVerificationResponse, error) {
	h := c.accountHeaders()
	h.Set("X-Captcha-Token", captchaToken)
	return request[SMSVerificationResponse](c, "POST", accountBase+"/v1/auth/verification", SMSVerificationRequest{
		PhoneNumber: phone, Target: target, ClientID: clientID,
	}, h)
}

// 验证短信验证码
// id 验证码ID
// code 验证码
func (c *Client) LoginSMSVerify(varifacationId, varifacationCode string) (SMSVerifyResponse, error) {
	return request[SMSVerifyResponse](c, "POST", accountBase+"/v1/auth/verification/verify", SMSVerifyRequest{
		VerificationID: varifacationId, VerificationCode: varifacationCode, ClientID: clientID,
	}, c.accountHeaders())
}

// 短信登录
// varifacationCode 验证码
// verificationToken 验证码token
// username 用户名
// captchaToken 验证码token
func (c *Client) LoginSMSSignin(varifacationCode, verificationToken, username, captchaToken string) (TokenResponse, error) {
	h := c.accountHeaders()
	h.Set("X-Captcha-Token", captchaToken)
	result, err := request[TokenResponse](c, "POST", accountBase+"/v1/auth/signin", SMSSigninRequest{
		VerificationCode: varifacationCode, VerificationToken: verificationToken, Username: username, ClientID: clientID,
	}, h)
	if err == nil {
		c.updateTokens(result)
	}
	return result, err
}

// 短信登录全流程
// phone 手机号
// getCode 获取验证码的回调函数
// target 目标
// err 错误
func (c *Client) LoginSMS(phone string, getCode func() (string, error), target string) (TokenResponse, error) {
	if target == "" {
		target = "ANY"
	}
	initResult, err := c.LoginSMSInit(phone, "")
	if err != nil {
		return TokenResponse{}, err
	}
	if initResult.CaptchaToken == "" {
		return TokenResponse{}, errors.New("login init response has no captcha_token")
	}
	sendResult, err := c.LoginSMSSend(phone, initResult.CaptchaToken, target)
	if err != nil {
		return TokenResponse{}, err
	}
	if sendResult.VerificationID == "" {
		return TokenResponse{}, errors.New("login send response has no verification_id")
	}
	if getCode == nil {
		return TokenResponse{}, errors.New("getCode callback is required")
	}
	code, err := getCode()
	if err != nil {
		return TokenResponse{}, err
	}
	verifyResult, err := c.LoginSMSVerify(sendResult.VerificationID, code)
	if err != nil {
		return TokenResponse{}, err
	}
	if verifyResult.VerificationToken == "" {
		return TokenResponse{}, errors.New("login verify response has no verification_token")
	}
	return c.LoginSMSSignin(code, verifyResult.VerificationToken, phone, initResult.CaptchaToken)
}

// 刷新token
// refresh 刷新token
// err 错误
func (c *Client) RefreshToken(refresh string) (TokenResponse, error) {
	if refresh == "" {
		refresh = c.RefreshTokenValue
	}
	if refresh == "" {
		return TokenResponse{}, errors.New("no refresh token available")
	}
	h := c.accountHeaders()
	h.Set("X-Action", "401")
	result, err := request[TokenResponse](c, "POST", accountBase+"/v1/auth/token", RefreshTokenRequest{
		ClientID: clientID, GrantType: "refresh_token", RefreshToken: refresh,
	}, h)
	if err != nil {
		return result, err
	}
	c.updateTokens(result)
	if c.RefreshTokenValue == "" {
		c.RefreshTokenValue = refresh
	}
	return result, err
}

// 查询扫码登录结果并换取token
// deviceCode 设备授权码
func (c *Client) DeviceToken(deviceCode string) (TokenResponse, error) {
	result, err := request[TokenResponse](c, "POST", accountBase+"/v1/auth/token", DeviceTokenRequest{
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		DeviceCode: deviceCode,
		ClientID:   clientID,
	}, c.accountHeaders())
	if err != nil {
		return result, err
	}
	c.updateTokens(result)
	return result, nil
}

// 获取扫码登录二维码
// scope 授权范围，默认：user
func (c *Client) DeviceCode(scope string) (DeviceCodeResponse, error) {
	return request[DeviceCodeResponse](c, "POST", accountBase+"/v1/auth/device/code", DeviceCodeRequest{
		Scope: scope, ClientID: clientID,
	}, c.accountHeaders())
}

// 更新token
// result token响应
func (c *Client) updateTokens(result TokenResponse) {
	if result.AccessToken != "" {
		c.Token = result.AccessToken
	}
	if result.ExpiresIn > 0 {
		c.TokenExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	}
	if result.RefreshToken != "" {
		c.RefreshTokenValue = result.RefreshToken
	}
}

// 获取用户信息
func (c *Client) UserInfo() (UserInfoResponse, error) {
	h := c.accountHeaders()
	h.Set("Authorization", "Bearer "+c.Token)
	return request[UserInfoResponse](c, "GET", accountBase+"/v1/user/me", nil, h)
}

// 获取用户资产
func (c *Client) GetAssets() (AssetsResponse, error) {
	return request[AssetsResponse](c, "POST", apiBase+"/assets/v1/get_assets", EmptyRequest{}, nil)
}

// 获取流量统计
// bizType 业务类型， 8：直链流量包，15： 免登录流量包
// groupBy 分组方式
// startDate 开始日期，格式为 YYYY-MM-DD
// endDate 结束日期，格式为 YYYY-MM-DD
func (c *Client) GetTrafficStatistics(bizType, groupBy int, startDate, endDate string) (TrafficStatisticsResponse, error) {
	return request[TrafficStatisticsResponse](c, "POST", apiBase+"/assets/v1/get_traffic_statistics", TrafficStatisticsRequest{
		BizType: bizType, GroupBy: groupBy, StartDate: startDate, EndDate: endDate,
	}, nil)
}

// 获取应用列表
// msgType 消息类型
// page 页码
// pageSize 每页数量
func (c *Client) GetInAppMsgList(msgType, page, pageSize int) (InAppMsgListResponse, error) {
	return request[InAppMsgListResponse](c, "POST", apiBase+"/misc/v1/get_inapp_msg_list", InAppMsgListRequest{
		MsgType: msgType, Page: page, PageSize: pageSize,
	}, nil)
}

// 获取用户操作记录列表
// pageSize 每页数量
// cursor 分页游标
// fileTypes 文件类型列表
func (c *Client) GetUserAction(pageSize int, cursor string, fileTypes []int) (UserActionResponse, error) {
	return request[UserActionResponse](c, "POST", apiBase+"/userres/v1/get_user_action", UserActionRequest{
		PageSize: pageSize, Cursor: cursor, FileTypes: fileTypes,
	}, nil)
}

// 获取转存文件列表
// pageSize 每页数量
// cursor 分页游标
// orderBy 排序字段
// sortType 排序类型
// err 错误
func (c *Client) GetRestoreList(pageSize, cursor, orderBy, sortType int) (RestoreListResponse, error) {
	return request[RestoreListResponse](c, "POST", apiBase+"/userres/v1/get_restore_list", RestoreListRequest{
		PageSize: pageSize, Cursor: cursor, OrderBy: orderBy, SortType: sortType,
	}, nil)
}

// 搜索文件
// name 文件名
// pageSize 每页数量
func (c *Client) SearchFiles(name string, pageSize int) (FileListResponse, error) {
	return request[FileListResponse](c, "POST", apiBase+"/userres/v1/file/search_files", SearchFilesRequest{
		Name: name, PageSize: pageSize,
	}, nil)
}

// 获取压缩文件列表
// fileID 压缩文件ID
// pageSize 每页数量
// password 压缩文件密码
func (c *Client) GetCompressFileList(fileID string, pageSize int, password string) (CompressFileListResponse, error) {
	return request[CompressFileListResponse](c, "POST", apiBase+"/userres/v1/get_compress_file_list", CompressFileListRequest{
		FileID: fileID, PageSize: pageSize, Password: password,
	}, nil)
}

// 解压缩文件
// fileID 压缩文件ID
// password 压缩文件密码
// filePaths 待解压的文件路径列表
// toFileID 解压目标目录ID
func (c *Client) DecompressFiles(fileID, password string, filePaths []string, toFileID string) (FileTaskResponse, error) {
	return request[FileTaskResponse](c, "POST", apiBase+"/userres/v1/decompress_files", DecompressFilesRequest{
		FileID: fileID, Password: password, FilePaths: filePaths, ToFileID: toFileID,
	}, nil)
}

// 查询解压缩任务状态
// taskID 解压缩任务ID
func (c *Client) QueryDecompressStatus(taskID string) (TaskStatusResponse, error) {
	return request[TaskStatusResponse](c, "POST", apiBase+"/userres/v1/query_decompress_status", QueryDecompressStatusRequest{
		TaskID: taskID,
	}, nil)
}

// 解析下载链接
// url HTTP/磁力/ed2k链接
func (c *Client) CloudResolveURL(url string) (CloudResolveURLResponse, error) {
	return request[CloudResolveURLResponse](c, "POST", apiBase+"/nd.bizcloudcollection.s/v1/resolve_res", CloudResolveURLRequest{URL: url}, nil)
}

// 创建云添加任务
// fileIndexes BT文件索引列表
// url HTTP/磁力/ed2k链接
// parentId 父目录ID
// newName 新文件名
func (c *Client) CloudCreateTask(fileIndexes []int, url string, parentId any, newName string) (CloudCreateTaskResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	return request[CloudCreateTaskResponse](c, "POST", apiBase+"/nd.bizcloudcollection.s/v1/create_task", CloudCreateTaskRequest{
		FileIndexes: fileIndexes, URL: url, ParentID: parentId, NewName: newName,
	}, nil)
}

// 获取下载URL
// fileId 文件ID
func (c *Client) DownloadURL(fileId string) (DownloadURLResponse, error) {
	return request[DownloadURLResponse](c, "POST", apiBase+"/userres/v1/get_res_download_url", FileIDRequest{FileID: fileId}, nil)
}

// 创建目录
// dirName 目录名
// parentId 父目录ID
// failIfNameExist 如果目录名已存在，是否返回失败，true时返回目录ID
// err 错误
func (c *Client) FSCreateDir(dirName string, parentId any, failIfNameExist bool) (FSCreateDirResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	p := FSCreateDirRequest{DirName: dirName, ParentID: parentId, FailIfNameExist: failIfNameExist}
	return request[FSCreateDirResponse](c, "POST", apiBase+"/userres/v1/file/create_dir", p, nil)
}

// 复制文件
// fileIds 文件ID列表
// parentId 父目录ID
func (c *Client) FSCopy(fileIds []any, parentId any) (FileTaskResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	return request[FileTaskResponse](c, "POST", apiBase+"/userres/v1/file/copy_file", FileIDsRequest{FileIDs: fileIds, ParentID: parentId}, nil)
}

// 移动文件
// fileIds 文件ID列表
// parentId 父目录ID
func (c *Client) FSMove(fileIds []any, parentId any) (FileTaskResponse, error) {
	if parentId == nil {
		parentId = ""
	}
	return request[FileTaskResponse](c, "POST", apiBase+"/userres/v1/file/move_file", FileIDsRequest{FileIDs: fileIds, ParentID: parentId}, nil)
}

// 删除文件
// fileIds 文件ID列表
func (c *Client) FSDelete(fileIds []any) (FileTaskResponse, error) {
	return request[FileTaskResponse](c, "POST", apiBase+"/userres/v1/file/delete_file", FileIDsRequest{FileIDs: fileIds}, nil)
}

// 重命名文件
// fileId 文件ID
// newName 新文件名
func (c *Client) FSRename(fileId string, newName string) (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/file/rename", FSRenameRequest{FileID: fileId, NewName: newName}, nil)
}

// 获取文件详情
// fileId 文件ID
func (c *Client) FSDetail(fileId string) (FileDetailResponse, error) {
	return request[FileDetailResponse](c, "POST", apiBase+"/userres/v1/file/get_file_detail", FileIDRequest{FileID: fileId}, nil)
}

// 获取图片列表
// page 页码
// size 每页数量
// order 排序字段
// sort 排序方式
func (c *Client) FSImageList(page, pageSize, orderBy, sortType int) (FileListResponse, error) {
	one := 1
	return c.FSFiles("*", page, pageSize, orderBy, sortType, []int{1}, &one, nil, false)
}

// 获取视频列表
// page 页码
// pageSize 每页数量
// orderBy 排序字段
// sortType 排序方式
// play 是否播放
func (c *Client) FSVideoList(page, pageSize, orderBy, sortType int, play bool) (FileListResponse, error) {
	one := 1
	return c.FSFiles("*", page, pageSize, orderBy, sortType, []int{2}, &one, nil, play)
}

// 获取文档列表
// page 页码
// pageSize 每页数量
// orderBy 排序字段
// sortType 排序方式
func (c *Client) FSDocumentList(page, pageSize, orderBy, sortType int) (FileListResponse, error) {
	one := 1
	return c.FSFiles("*", page, pageSize, orderBy, sortType, []int{4}, &one, nil, false)
}

// 获取回收站文件列表
// page 页码
// pageSize 每页数量
// orderBy 排序字段
// sortType 排序方式
func (c *Client) FSRecycleFiles(page, pageSize, orderBy, sortType int) (FileListResponse, error) {
	four := 4
	return c.FSFiles(nil, page, pageSize, orderBy, sortType, nil, nil, &four, false)
}

// 恢复回收站文件
// ids 文件ID列表
func (c *Client) FSRecycle(fileIds []any) (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/file/recycle_file", FileIDsRequest{FileIDs: fileIds}, nil)
}

// 清空回收站
func (c *Client) FSClearRecycleBin() (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/file/clear_recycle_bin", EmptyRequest{}, nil)
}

// 分享文件
// params 完整分享参数，按原值发送，不填充默认值
func (c *Client) ShareCreate(params ShareCreateRequest) (ShareCreateResponse, error) {
	return request[ShareCreateResponse](c, "POST", apiBase+"/userres/v1/share_file", params, nil)
}

// 更新分享
// shareID 分享ID
// title 分享标题
// validateDuration 验证时间
// shareType 分享类型
// code 分享码
// autoFillCode 是否自动填充码
// trafficLimit 流量限制
// maxRestoreCount 最大转存次数
// downloadType 下载类型
func (c *Client) ShareUpdate(id, title string, validateDuration, shareType int, code string, autoFillCode bool, trafficLimit string, maxRestoreCount, downloadType int) (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/update_share", ShareUpdateRequest{
		ID: id, Title: title, ValidateDuration: validateDuration, ShareType: shareType, Code: code,
		AutoFillCode: autoFillCode, TrafficLimit: trafficLimit, MaxRestoreCount: maxRestoreCount,
		DownloadType: downloadType,
	}, nil)
}

// 获取分享列表
// page 页码
// pageSize 每页数量
// orderBy 排序字段
// sortType 排序方式
func (c *Client) ShareUserList(page, pageSize, orderBy, sortType int) (ShareListResponse, error) {
	return request[ShareListResponse](c, "POST", apiBase+"/userres/v1/get_share_list", ShareListRequest{
		Page: page, PageSize: pageSize, OrderType: orderBy, SortType: sortType,
	}, nil)
}

// ShareAuditRejectList 获取分享中审核未通过的文件列表，参数按原值发送。
func (c *Client) ShareAuditRejectList(params ShareAuditRejectListRequest) (ShareAuditRejectListResponse, error) {
	return request[ShareAuditRejectListResponse](c, "POST", apiBase+"/userres/v1/get_share_audit_reject_list", params, nil)
}

// 删除分享
// ids 分享ID列表
func (c *Client) ShareDelete(ids []any) (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/delete_share", ShareDeleteRequest{IDs: ids}, nil)
}

// ShareDeleteInvalid 删除失效分享。
func (c *Client) ShareDeleteInvalid() (EmptyResponse, error) {
	return request[EmptyResponse](c, "POST", apiBase+"/userres/v1/delete_invalid_share", EmptyRequest{}, nil)
}

// 转存分享
// accessToken 分享访问令牌
// fileIds 要转存的文件ID列表
// parentId 目标目录ID
func (c *Client) ShareRestore(accessToken string, fileIds []any, parentId string) (FileTaskResponse, error) {
	return request[FileTaskResponse](c, "POST", apiBase+"/userres/v1/restore_share", ShareRestoreRequest{
		AccessToken: accessToken, FileIDs: fileIds, ParentID: parentId,
	}, nil)
}

// 获取任务状态，删除、复制、移动、转存任务状态
// taskId 任务ID
func (c *Client) GetTaskStatus(taskId string) (TaskStatusResponse, error) {
	return request[TaskStatusResponse](c, "POST", apiBase+"/userres/v1/get_task_status", TaskStatusRequest{TaskID: taskId}, nil)
}

// 获取分享下载URL
// fileId 分享文件ID
// accessToken 分享访问令牌
func (c *Client) ShareDownloadURL(fileId, accessToken string) (ShareDownloadURLResponse, error) {
	return request[ShareDownloadURLResponse](c, "POST", apiBase+"/userres/v1/get_share_download_url", ShareDownloadURLRequest{
		FileID: fileId, AccessToken: accessToken,
	}, nil)
}

// 获取分享文件大小
// token 分享访问令牌
// ids 要获取大小的文件ID列表
// download 是否下载文件
func (c *Client) ShareFilesSize(accessToken string, fileIds []any, download bool) (ShareFilesSizeResponse, error) {
	return request[ShareFilesSizeResponse](c, "POST", apiBase+"/userres/v1/get_share_files_size", ShareFilesSizeRequest{
		AccessToken: accessToken, FileIDs: fileIds, Download: download,
	}, nil)
}

// 获取分享摘要
// shareId 分享ID
func (c *Client) ShareSummary(shareId string) (ShareSummaryResponse, error) {
	return publicPost[ShareSummaryResponse](c, "/userres/v1/get_share_summary", ShareIDRequest{ShareID: shareId})
}

// 获取分享访问令牌
// shareId 分享ID
// code 分享码
func (c *Client) ShareAccessToken(shareId, code string) (ShareAccessTokenResponse, error) {
	return request[ShareAccessTokenResponse](c, "POST", apiBase+"/userres/v1/get_share_access_token", ShareAccessTokenRequest{
		ShareID: shareId, Code: code,
	}, nil)
}

// 获取分享文件列表
// accessToken 分享访问令牌
// parentId 目录ID
// page 页码
// pageSize 每页数量
// orderBy 排序字段
// sortType 排序方式
func (c *Client) ShareFilesList(accessToken string, parentId string, page, pageSize, orderBy, sortType int) (ShareFilesListResponse, error) {
	return publicPost[ShareFilesListResponse](c, "/userres/v1/get_share_page_files_list", ShareFilesListRequest{
		AccessToken: accessToken, ParentID: parentId, Page: page, PageSize: pageSize, OrderBy: orderBy, SortType: sortType,
	})
}

type ossComplete struct {
	ETag string `xml:"ETag"`
}
type ossInit struct {
	UploadID string `xml:"UploadId"`
}

func OSSSignHeaders(method, bucket, key, accessKey, secret, securityToken, contentType, contentMD5 string, sub map[string]string) http.Header {
	date := time.Now().UTC().Format(http.TimeFormat)
	canonicalHeaders := map[string]string{"x-oss-date": date, "x-oss-security-token": strings.TrimSpace(securityToken)}
	resource := "/" + bucket + "/" + key
	if len(sub) > 0 {
		keys := make([]string, 0, len(sub))
		for k := range sub {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			if sub[k] == "" {
				parts = append(parts, k)
			} else {
				parts = append(parts, k+"="+sub[k])
			}
		}
		resource += "?" + strings.Join(parts, "&")
	}
	lines := []string{strings.ToUpper(method), contentMD5, contentType, date, "x-oss-date:" + canonicalHeaders["x-oss-date"], "x-oss-security-token:" + canonicalHeaders["x-oss-security-token"], resource}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(strings.Join(lines, "\n")))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	h := http.Header{"Authorization": {"OSS " + accessKey + ":" + sig}, "X-Oss-Date": {date}, "X-Oss-Security-Token": {securityToken}}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	if contentMD5 != "" {
		h.Set("Content-MD5", contentMD5)
	}
	return h
}

func (c *Client) OSSRequest(method, rawURL, bucket, objectKey, accessKey, secret, securityToken string, content []byte, contentType, contentMD5 string, sub map[string]string) (*http.Response, error) {
	headers := OSSSignHeaders(method, bucket, objectKey, accessKey, secret, securityToken, contentType, contentMD5, sub)
	query := url.Values{}
	for key, value := range sub {
		query.Set(key, value)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	// OSS has its own credentials and signed headers; do must not replace them.
	req.Header = headers
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("guangya: OSS HTTP %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return resp, nil
}

func (c *Client) CDNUpload(path string, tokenData UploadTokenData, contentType string, chunkSize int64) (string, error) {
	if tokenData.Creds.AccessKeyID == "" || tokenData.Creds.SecretAccessKey == "" || tokenData.Creds.SessionToken == "" {
		return "", errors.New("upload token response has incomplete creds")
	}
	if tokenData.FullEndpoint == "" || tokenData.BucketName == "" || tokenData.ObjectPath == "" {
		return "", errors.New("upload token response has incomplete upload fields")
	}
	accessKey, secret, sessionToken := tokenData.Creds.AccessKeyID, tokenData.Creds.SecretAccessKey, tokenData.Creds.SessionToken
	endpoint, bucket, objectKey := tokenData.FullEndpoint, tokenData.BucketName, tokenData.ObjectPath
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024
	}
	baseURL := strings.TrimRight(endpoint, "/") + "/" + objectKey
	initial, err := c.OSSRequest("POST", baseURL, bucket, objectKey, accessKey, secret, sessionToken, nil, contentType, "", map[string]string{"uploads": ""})
	if err != nil {
		return "", err
	}
	var initResult ossInit
	if err := xml.NewDecoder(initial.Body).Decode(&initResult); err != nil {
		initial.Body.Close()
		return "", err
	}
	initial.Body.Close()
	if initResult.UploadID == "" {
		return "", errors.New("OSS response has no upload id")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	type part struct {
		Number int
		ETag   string
	}
	var parts []part
	buffer := make([]byte, chunkSize)
	for number := 1; ; number++ {
		n, readErr := io.ReadFull(file, buffer)
		if readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			return "", readErr
		}
		chunk := buffer[:n]
		sum := md5.Sum(chunk)
		contentMD5Value := base64.StdEncoding.EncodeToString(sum[:])
		response, requestErr := c.OSSRequest("PUT", baseURL, bucket, objectKey, accessKey, secret, sessionToken, chunk, contentType, contentMD5Value, map[string]string{"partNumber": strconv.Itoa(number), "uploadId": initResult.UploadID})
		if requestErr != nil {
			return "", requestErr
		}
		etag := strings.Trim(response.Header.Get("ETag"), `"`)
		response.Body.Close()
		if etag == "" {
			return "", errors.New("OSS upload part response has no ETag")
		}
		parts = append(parts, part{Number: number, ETag: etag})
		if readErr == io.ErrUnexpectedEOF {
			break
		}
	}
	var complete bytes.Buffer
	complete.WriteString(`<?xml version="1.0" encoding="UTF-8"?><CompleteMultipartUpload>`)
	for _, item := range parts {
		fmt.Fprintf(&complete, `<Part><PartNumber>%d</PartNumber><ETag>"%s"</ETag></Part>`, item.Number, item.ETag)
	}
	complete.WriteString(`</CompleteMultipartUpload>`)
	sum := md5.Sum(complete.Bytes())
	response, err := c.OSSRequest("POST", baseURL, bucket, objectKey, accessKey, secret, sessionToken, complete.Bytes(), "application/xml", base64.StdEncoding.EncodeToString(sum[:]), map[string]string{"uploadId": initResult.UploadID})
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var result ossComplete
	if err := xml.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.ETag == "" {
		return "", errors.New("OSS complete response has no ETag")
	}
	return result.ETag, nil
}

// 获取文件上传信息
// taskID 上传任务ID
func (c *Client) UploadInfo(taskID string) (UploadInfoResponse, error) {
	return request[UploadInfoResponse](c, "POST", apiBase+"/userres/v1/file/get_info_by_task_id", TaskStatusRequest{TaskID: taskID}, nil)
}

// 上传文件
// path 文件路径
// name 文件名
// parent 父目录ID
func (c *Client) FileUpload(path, name string, parentId any) (UploadInfoResponse, error) {
	info, err := os.Stat(path)
	if err != nil {
		return UploadInfoResponse{}, err
	}
	if name == "" {
		name = filepath.Base(path)
	}
	var md5Value string
	if info.Size() < 1024*1024 {
		data, err := os.ReadFile(path)
		if err != nil {
			return UploadInfoResponse{}, err
		}
		sum := md5.Sum(data)
		md5Value = hex.EncodeToString(sum[:])
	}
	token, err := c.UploadToken(name, info.Size(), parentId, md5Value)
	fmt.Println("token:", token)
	if err != nil {
		return UploadInfoResponse{}, err
	}
	if token.Code != 0 {
		return UploadInfoResponse{}, fmt.Errorf("guangya: upload token code %d: %s", token.Code, token.Msg)
	}
	data := token.Data
	task := data.TaskID
	if task == "" {
		return UploadInfoResponse{}, errors.New("upload token response has no taskId")
	}
	flash, err := c.CheckCanFlashUpload(task, path)
	fmt.Println("flash:", flash)
	if err != nil {
		return UploadInfoResponse{}, err
	}
	if flash.Code != 0 {
		return UploadInfoResponse{}, fmt.Errorf("guangya: flash upload code %d: %s", flash.Code, flash.Msg)
	}
	if flash.Data.CanFlashUpload {
		return c.waitUpload(task, 4, 5*time.Second)
	}
	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if _, err := c.CDNUpload(path, data, contentType, 5*1024*1024); err != nil {
		return UploadInfoResponse{}, err
	}
	return c.waitUpload(task, 4, 5*time.Second)
}

// 等待文件上传完成
// task 上传任务ID
// attempts 尝试次数
// delay 每次检查间隔
func (c *Client) waitUpload(task string, attempts int, delay time.Duration) (UploadInfoResponse, error) {
	var result UploadInfoResponse
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
		}
		var err error
		result, err = c.UploadInfo(task)
		if err != nil {
			return result, err
		}
		if result.Code == 147 || result.Msg == "文件上传中" {
			continue
		}
		if result.Code != 0 {
			return result, fmt.Errorf("guangya: upload task %s code %d: %s", task, result.Code, result.Msg)
		}
		return result, nil
	}
	return result, fmt.Errorf("guangya: upload task %s still pending after %d checks", task, attempts)
}

// 解析BT种子
// data BT种子数据
// filename BT种子文件名
func (c *Client) ResolveTorrent(data []byte, filename string) (CloudResolveTorrentResponse, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("torrent", filename)
	if err != nil {
		return CloudResolveTorrentResponse{}, err
	}
	if _, err = part.Write(data); err != nil {
		return CloudResolveTorrentResponse{}, err
	}
	if err := mw.Close(); err != nil {
		return CloudResolveTorrentResponse{}, err
	}
	h := http.Header{"Content-Type": {mw.FormDataContentType()}}
	resp, err := c.do("POST", apiBase+"/nd.bizcloudcollection.s/v1/resolve_torrent", body.Bytes(), h, nil)
	if err != nil {
		return CloudResolveTorrentResponse{}, err
	}
	defer resp.Body.Close()
	var out CloudResolveTorrentResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out, err
}

var _ = xml.Name{}
var _ = strconv.Itoa

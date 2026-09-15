# guangyaclient-go

接口文档：[接口列表](#api-list) · [请求参数](#api-requests) · [响应结构](#api-responses) · [新增接口示例](#api-examples)

光鸭云盘客户端的 Go 实现，基于 Go 标准库，无第三方依赖。

## 安装

在项目中执行：

```bash
go get github.com/BugGeeker/guangyaclient-go
```

导入包：

```go
import guangyaclient "github.com/BugGeeker/guangyaclient-go"
```

## 初始化客户端

```go
client := guangyaclient.NewClient(
	"access-token",  // 没有时传空字符串
	"refresh-token", // 没有时传空字符串
	"",              // 设备 ID，传空则自动生成
)
defer client.Close()
```

建议保存 `Token` 和 `RefreshTokenValue`，下次启动时传给 `NewClient`，避免重复登录。

## 短信登录

短信登录需要提供一个获取验证码的回调：

```go
client := guangyaclient.NewClient("", "", "")

result, err := client.LoginSMS("+86 13800138000", func() (string, error) {
	fmt.Print("请输入短信验证码: ")
	var code string
	_, err := fmt.Scanln(&code)
	return code, err
}, "ANY")
if err != nil {
	log.Fatal(err)
}

fmt.Println("登录结果:", result)
fmt.Println("Access Token:", client.Token)
fmt.Println("Refresh Token:", client.RefreshTokenValue)
```

刷新 Token：

```go
result, err := client.RefreshToken("")
```

传入空字符串时使用客户端保存的 `RefreshTokenValue`。普通 API 请求会在 Token 过期或收到 HTTP 401 时自动刷新并重试。

## 获取用户信息

```go
user, err := client.UserInfo()
if err != nil {
	log.Fatal(err)
}
fmt.Println(user.Subject, user.Name, user.PhoneNumber)
fmt.Println(user.CreatedAt, user.PasswordUpdatedAt)
```

## 获取账户资产

`GetAssets()` 无入参，返回 `AssetsResponse`。容量、流量和时间字段使用 `int64`，会员状态使用 `int`。

```go
assets, err := client.GetAssets()
if err != nil {
	log.Fatal(err)
}
fmt.Println(assets.Code, assets.Msg)
fmt.Println(assets.Data.TotalSpaceSize, assets.Data.UsedSpaceSize)
fmt.Println(assets.Data.TotalDirectLinkTraffic, assets.Data.FreeDirectLinkTraffic)
fmt.Println(assets.Data.TotalShareGuestTraffic, assets.Data.FreeShareGuestTraffic)
fmt.Println(assets.Data.VIPStatus, assets.Data.SVIPStatus, assets.Data.VIPLeftTime)
fmt.Println(assets.Data.VIPExpireTime, assets.Data.SystemTime)
```

## 文件管理

文件列表、文件详情、下载地址、分享创建、云任务、上传凭证等接口返回强类型结构：

```go
// 获取根目录文件
files, err := client.FSFiles(
	nil, // parentID
	0,   // page
	50,  // pageSize
	0,   // orderBy
	0,   // sortType
	nil, // fileTypes
	nil, // resType
	nil, // dirType
	false,
)

// 创建目录
created, err := client.FSCreateDir("新目录", nil, false)
fmt.Println(created.Data.FileID, created.Data.FileName, created.Data.ParentID)

// 重命名文件
renamed, err := client.FSRename("file-id", "new-name.txt")

// 删除文件
deleted, err := client.FSDelete([]any{"file-id-1", "file-id-2"})
fmt.Println(deleted.Data.TaskID)
```

按类型获取文件：

```go
images, _ := client.FSImageList(0, 50, 3, 1)
videos, _ := client.FSVideoList(0, 50, 3, 1, true)
documents, _ := client.FSDocumentList(0, 50, 3, 1)
recycleBin, _ := client.FSRecycleFiles(0, 50, 10, 0)
```

文件类型常量：

```go
fmt.Println(guangyaclient.FileType["图片"])       // 1
fmt.Println(guangyaclient.FileTypeName[1])       // 图片
```

`FSCopy`、`FSMove` 返回 `FileTaskResponse`，可通过 `result.Data.TaskID`
获取任务 ID。`FSRecycle`、`FSClearRecycleBin` 返回 `EmptyResponse`，
日志中的 `data: null` 对应 `result.Data == nil`。

未完成 schema 建模的接口返回 `GenericResponse`，其中 `Data` 为
`json.RawMessage`，不会通过 `map[string]any` 丢失大整数精度。

`GetTaskStatus` 和 `QueryDecompressStatus` 返回 `TaskStatusResponse`，可通过 `Data.Status` 读取整数状态，通过 `Data.Progress` 读取进度：

```go
result, err := client.GetTaskStatus("task-id")
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.Msg, result.Data.Status)
```

## 下载与云添加

```go
// 获取文件下载地址
download, err := client.DownloadURL("file-id")
fmt.Println(download.Data.SignedURL, download.Data.URLDuration)

// 创建云添加任务
task, err := client.CloudCreateTask([]int{82, 83}, "magnet:?xt=urn:btih:65D1A45E6CB2A392A55E877DBA6BA603B91C1579", "1893227021252935776", "kpkp69.com-ROYD036")

// 查询云添加任务
tasks, err := client.CloudTaskList(0, 50, nil)

// 解析 HTTP、磁力或 ed2k 链接
resolved, err := client.CloudResolveURL("magnet:?xt=urn:btih:...")
fmt.Println(resolved.Data.BTResInfo.InfoHash, resolved.Data.BTResInfo.SubfilesNum)

// 解析本地种子文件，返回 CloudResolveTorrentResponse
torrentBytes, err := os.ReadFile("./example.torrent")
if err != nil {
	log.Fatal(err)
}
torrent, err := client.ResolveTorrent(torrentBytes, "example.torrent")
if err != nil {
	log.Fatal(err)
}
fmt.Println(torrent.Data.BTResInfo.InfoHash, torrent.Data.BTResInfo.FileSize)

// 查询任务状态
status, err := client.GetTaskStatus("task-id")
fmt.Println(status.Data.Status)
```

## 分享

```go
// 创建分享
share, err := client.ShareCreate(
	[]any{"file-id"},
	"我的文件",
)
fmt.Println(share.Data.ShareURL, share.Data.Code)

// 查询自己的分享
shares, err := client.ShareUserList(0, 50, 1, 1)

// 获取公开分享摘要，不要求登录
summary, err := client.ShareSummary("share-id")
fmt.Println(summary.Data.Title, summary.Data.NeedCode, summary.Data.TotalFileSize)

// 获取分享访问令牌，不要求登录
access, err := client.ShareAccessToken("share-id", "提取码")

// 查询分享文件
sharedFiles, err := client.ShareFilesList("access-token", "", 1, 50, 0, 0)
fmt.Println(sharedFiles.Data.Total, sharedFiles.Data.Cursor)
for _, file := range sharedFiles.Data.List {
	fmt.Println(file.FileID, file.FileName, file.FileSize)
}

// 查询分享文件大小和下载地址
sizes, err := client.ShareFilesSize("access-token", []any{"file-id"}, true)
fmt.Println(sizes.Data.TotalSize, sizes.Data.FileSizeMap["file-id"])
download, err := client.ShareDownloadURL("file-id", "access-token")
fmt.Println(download.Data.DownloadURL)

// 转存分享，返回 FileTaskResponse
restored, err := client.ShareRestore("access-token", []any{"file-id"}, "")
if err != nil {
	log.Fatal(err)
}
fmt.Println(restored.Data.TaskID)
```

`ShareSummary` 返回 `ShareSummaryResponse`，`ShareFilesList` 返回
`ShareFilesListResponse`（文件条目复用 `FileItem`，`Cursor` 为整数）。
`ShareUpdate`、`ShareDelete` 返回 `EmptyResponse`（`data: null` 对应 `Data == nil`），`ShareFilesSize` 返回
`ShareFilesSizeResponse`，`ShareDownloadURL` 返回 `ShareDownloadURLResponse`。

## 文件上传

小于 1 MB 的文件申请上传凭证时会附带十六进制 MD5。所有文件都会执行
GCID/CID 秒传检查；未命中秒传时继续执行 OSS 分片上传，完成后轮询文件信息。
返回 `code: 147` 表示仍在处理，客户端最多查询 4 次、间隔 5 秒；
超时会返回错误，可使用错误中的任务 ID 调用 `UploadInfo` 继续查询。

`UploadInfo` 和 `FileUpload` 均返回 `UploadInfoResponse`，其中 `Data` 为
`*FileDetailInfo`，可读取 `FileID`、`FileName`、`FileSize`、`MD5` 等字段。
上传中响应的数据可能为空，访问字段前需检查 `result.Data != nil`。

```go
result, err := client.FileUpload(
	"./movie.mp4",
	"",  // 文件名，空字符串使用本地文件名
	nil, // 父目录 ID
)
if err != nil {
	log.Fatal(err)
}
fmt.Println(result)
```

也可以单独申请上传凭证：

```go
token, err := client.UploadToken("movie.mp4", fileSize, nil, "")
```

上传凭证成功响应可以不包含 `code`，对应 `token.Code == 0`。
`token.Data` 保留 `GCID`、`Provider`、`Endpoint`、`CallbackVar`、`Region` 等字段；
`token.Data.Creds.Expiration` 为可选的 `*time.Time`，缺省时为 `nil`。

## 错误处理

```go
result, err := client.FSDetail("file-id")
if err != nil {
	var apiError error
	if errors.As(err, &apiError) {
		log.Println("请求失败:", apiError)
	}
	return
}
fmt.Println(result)
```

请求失败时会返回包含 HTTP 状态码和响应内容的 error。不要忽略 API 方法返回的 `error`。

<a id="api-list"></a>

## 接口列表

以下内容以当前 `client.go`、`request.go` 和 `types.go` 的实现为准，描述 SDK 实际发送的字段和已建模的响应，不代表服务端完整协议。

### 通用约定

- 账户接口的基础地址为 `https://account.guangyapan.com`；其他业务接口的基础地址为 `https://api.guangyapan.com`。下表中的路径均为相对于所属基础地址的路径。
- 所有业务方法均返回 `(响应类型, error)`。普通 POST 接口使用 JSON；`ResolveTorrent` 使用 `multipart/form-data`；OSS 使用独立签名、二进制或 XML 请求。
- 请求参数表中的名称是 **JSON 字段名**，类型是 Go DTO 中的类型。实际调用仍使用接口列表中所示的方法参数，不是直接将 DTO 传入方法。
- “省略”指当前 SDK 的 JSON 序列化行为，不表示已验证服务端的必填规则。没有注明省略的字段，即使值为 `0`、`false`、`""` 或 `nil`，也会发送。
- `any`、`[]any` 类型的 ID 参数保留兼容性，按传入值序列化。建议大整数 ID 使用字符串，避免调用方转换时丢失精度。
- 分页起点、排序枚举、数量上限等未在 SDK 中校验或统一定义。除明确列出的默认值外，SDK 原样传递参数，不自动翻页。
- `GetUserAction` 的 `cursor` 是字符串；`GetRestoreList` 的 `cursor` 是整数；`ShareFilesList` 的响应也有整数 `cursor`，但请求使用 `page`，不能混用。
- 认证请求的 `client_id` 和账户请求头 `X-Client-Id` 由 `const.go` 的 `clientID` 提供。账户请求头还使用客户端的 `DeviceID`。
- 普通业务请求由公共请求层补充 `Authorization: Bearer <Token>`、`DID`、`DT`、`Traceparent` 等请求头。公开分享封装也经过该请求层；是否允许空账户 Token 由服务端决定。分享参数 `accessToken` 与账户 `client.Token` 是不同的令牌。

### 账户认证

下表均使用账户基础地址。

| 方法及参数顺序                                                                              | HTTP 与路径                            | 请求类型                     | 返回类型                      |
| ------------------------------------------------------------------------------------ | ----------------------------------- | ------------------------ | ------------------------- |
| `LoginSMSInit(phone, captchaToken string)`                                           | `POST /v1/shield/captcha/init`      | `CaptchaInitRequest`     | `CaptchaInitResponse`     |
| `LoginSMSSend(phone, captchaToken, target string)`                                   | `POST /v1/auth/verification`        | `SMSVerificationRequest` | `SMSVerificationResponse` |
| `LoginSMSVerify(varifacationId, varifacationCode string)`                            | `POST /v1/auth/verification/verify` | `SMSVerifyRequest`       | `SMSVerifyResponse`       |
| `LoginSMSSignin(varifacationCode, verificationToken, username, captchaToken string)` | `POST /v1/auth/signin`              | `SMSSigninRequest`       | `TokenResponse`           |
| `RefreshToken(refresh string)`                                                       | `POST /v1/auth/token`               | `RefreshTokenRequest`    | `TokenResponse`           |
| `DeviceToken(deviceCode string)`                                                     | `POST /v1/auth/token`               | `DeviceTokenRequest`     | `TokenResponse`           |
| `UserInfo()`                                                                         | `GET /v1/user/me`                   | 无业务参数                    | `UserInfoResponse`        |

账户响应直接位于 JSON 顶层，不使用通用 `data` 包装。`UserInfo()` 当前复用 `request(..., nil, ...)`，会发送序列化后的 `null`，而非业务参数对象。

`DeviceCode(scope string)` 调用 `POST /v1/auth/device/code`，请求体为：

```json
{
  "scope": "user",
  "client_id": "aMe-8VSlkrbQXpUR"
}
```

返回类型为 `DeviceCodeResponse`，字段包括 `device_code`、`user_code`、`expires_in`、`interval`、`verification_url` 和 `verification_uri_complete`。

扫码确认后，可使用 `DeviceToken(deviceCode string)` 查询扫码结果并换取账户令牌。该方法同样调用 `POST /v1/auth/token`，请求体为：

```json
{
  "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
  "device_code": "AWqnqgTc2qJRvWz2sxmAPSPt2vsOWYm6DExC76hXLSwYT5HjyA",
  "client_id": "aMe-8VSlkrbQXpUR"
}
```

成功后返回 `TokenResponse`，并自动更新客户端访问令牌、刷新令牌和过期时间。

### 资产、消息与记录

| 方法及参数顺序                                                                 | HTTP 与路径                                 | 请求类型                       | 返回类型                        |
| ----------------------------------------------------------------------- | ---------------------------------------- | -------------------------- | --------------------------- |
| `GetAssets()`                                                           | `POST /assets/v1/get_assets`             | `EmptyRequest`             | `AssetsResponse`            |
| `GetTrafficStatistics(bizType, groupBy int, startDate, endDate string)` | `POST /assets/v1/get_traffic_statistics` | `TrafficStatisticsRequest` | `TrafficStatisticsResponse` |
| `GetInAppMsgList(msgType, page, pageSize int)`                          | `POST /misc/v1/get_inapp_msg_list`       | `InAppMsgListRequest`      | `InAppMsgListResponse`      |
| `GetUserAction(pageSize int, cursor string, fileTypes []int)`           | `POST /userres/v1/get_user_action`       | `UserActionRequest`        | `UserActionResponse`        |
| `GetRestoreList(pageSize, cursor, orderBy, sortType int)`               | `POST /userres/v1/get_restore_list`      | `RestoreListRequest`       | `RestoreListResponse`       |

### 文件管理与下载

| 方法及参数顺序                                                                                                                     | HTTP 与路径                                          | 请求类型                      | 返回类型                       |
| --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- | ------------------------- | -------------------------- |
| `FSFiles(parentId any, page, pageSize, orderBy, sortType int, fileTypes []int, resType, dirType *int, needPlayRecord bool)` | `POST /userres/v1/file/get_file_list`             | `FSFilesRequest`          | `FileListResponse`         |
| `SearchFiles(name string, pageSize int)`                                                                                    | `POST /userres/v1/file/search_files`              | `SearchFilesRequest`      | `FileListResponse`         |
| `GetCompressFileList(fileID string, pageSize int, password string)`                                                         | `POST /userres/v1/get_compress_file_list`         | `CompressFileListRequest` | `CompressFileListResponse` |
| `DecompressFiles(fileID, password string, filePaths []string, toFileID string)`                                            | `POST /userres/v1/decompress_files`              | `DecompressFilesRequest`  | `FileTaskResponse`         |
| `QueryDecompressStatus(taskID string)`                                                                                      | `POST /userres/v1/query_decompress_status`       | `QueryDecompressStatusRequest` | `TaskStatusResponse`       |
| `FSCreateDir(dirName string, parentId any, failIfNameExist bool)`                                                           | `POST /nd.bizuserres.s/v1/file/create_dir`        | `FSCreateDirRequest`      | `FSCreateDirResponse`      |
| `FSCopy(fileIds []any, parentId any)`                                                                                       | `POST /nd.bizuserres.s/v1/file/copy_file`         | `FileIDsRequest`          | `FileTaskResponse`         |
| `FSMove(fileIds []any, parentId any)`                                                                                       | `POST /nd.bizuserres.s/v1/file/move_file`         | `FileIDsRequest`          | `FileTaskResponse`         |
| `FSDelete(fileIds []any)`                                                                                                   | `POST /nd.bizuserres.s/v1/file/delete_file`       | `FileIDsRequest`          | `FileTaskResponse`         |
| `FSRename(fileId, newName string)`                                                                                          | `POST /nd.bizuserres.s/v1/file/rename`            | `FSRenameRequest`         | `EmptyResponse`            |
| `FSDetail(fileId string)`                                                                                                   | `POST /nd.bizuserres.s/v1/file/get_file_detail`   | `FileIDRequest`           | `FileDetailResponse`       |
| `FSRecycle(fileIds []any)`                                                                                                  | `POST /nd.bizuserres.s/v1/file/recycle_file`      | `FileIDsRequest`          | `EmptyResponse`            |
| `FSClearRecycleBin()`                                                                                                       | `POST /nd.bizuserres.s/v1/file/clear_recycle_bin` | `EmptyRequest`            | `EmptyResponse`            |
| `DownloadURL(fileId string)`                                                                                                | `POST /nd.bizuserres.s/v1/get_res_download_url`   | `FileIDRequest`           | `DownloadURLResponse`      |
| `GetTaskStatus(taskId string)`                                                                                              | `POST /nd.bizuserres.s/v1/get_task_status`        | `TaskStatusRequest`       | `TaskStatusResponse`       |

### 云添加与上传

| 方法及参数顺序                                                              | HTTP 与路径                                           | 请求类型                         | 返回类型                          |
| -------------------------------------------------------------------- | -------------------------------------------------- | ---------------------------- | ----------------------------- |
| `CloudResolveURL(url string)`                                        | `POST /nd.bizcloudcollection.s/v1/resolve_res`     | `CloudResolveURLRequest`     | `CloudResolveURLResponse`     |
| `CloudCreateTask(fileIndexes []int, url string, parentId any, newName string)` | `POST /nd.bizcloudcollection.s/v1/create_task` | `CloudCreateTaskRequest` | `CloudCreateTaskResponse` |
| `CloudTaskList(page, pageSize int, status []int)`                    | `POST /nd.bizcloudcollection.s/v1/list_task`       | `CloudTaskListRequest`       | `CloudTaskListResponse`       |
| `ResolveTorrent(data []byte, filename string)`                       | `POST /nd.bizcloudcollection.s/v1/resolve_torrent` | Multipart 文件字段 `torrent`     | `CloudResolveTorrentResponse` |
| `UploadToken(name string, size int64, parentId any, fileMD5 string)` | `POST /userres/v2/get_res_center_token`            | `UploadTokenRequest`         | `UploadTokenResponse`         |
| `CheckCanFlashUpload(taskID, path string)`                           | `POST /userres/v1/check_can_flash_upload`          | `CheckCanFlashUploadRequest` | `FlashUploadResponse`         |
| `UploadInfo(taskID string)`                                          | `POST /userres/v1/file/get_info_by_task_id`        | `TaskStatusRequest`          | `UploadInfoResponse`          |

### 分享管理

| 方法及参数顺序                                                                                                                                                  | HTTP 与路径                                             | 请求类型                      | 返回类型                       |
| -------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ------------------------- | -------------------------- |
| `ShareCreate(fileIds []any, title string)`                                                                                                               | `POST /nd.bizuserres.s/v1/share_file`                | `ShareCreateRequest`      | `ShareCreateResponse`      |
| `ShareUpdate(id, title string, validateDuration, shareType int, code string, autoFillCode bool, trafficLimit string, maxRestoreCount, downloadType int)` | `POST /nd.bizuserres.s/v1/update_share`              | `ShareUpdateRequest`      | `EmptyResponse`            |
| `ShareUserList(page, pageSize, orderBy, sortType int)`                                                                                                   | `POST /nd.bizuserres.s/v1/get_share_list`            | `ShareListRequest`        | `ShareListResponse`        |
| `ShareDelete(ids []any)`                                                                                                                                 | `POST /nd.bizuserres.s/v1/delete_share`              | `ShareDeleteRequest`      | `EmptyResponse`            |
| `ShareRestore(accessToken string, fileIds []any, parentId string)`                                                                                       | `POST /nd.bizuserres.s/v1/restore_share`             | `ShareRestoreRequest`     | `FileTaskResponse`         |
| `ShareSummary(shareId string)`                                                                                                                           | `POST /nd.bizuserres.s/v1/get_share_summary`         | `ShareIDRequest`          | `ShareSummaryResponse`     |
| `ShareAccessToken(shareId, code string)`                                                                                                                 | `POST /nd.bizuserres.s/v1/get_share_access_token`    | `ShareAccessTokenRequest` | `ShareAccessTokenResponse` |
| `ShareFilesList(accessToken, parentId string, page, pageSize, orderBy, sortType int)`                                                                    | `POST /nd.bizuserres.s/v1/get_share_page_files_list` | `ShareFilesListRequest`   | `ShareFilesListResponse`   |
| `ShareFilesSize(accessToken string, fileIds []any, download bool)`                                                                                       | `POST /nd.bizuserres.s/v1/get_share_files_size`      | `ShareFilesSizeRequest`   | `ShareFilesSizeResponse`   |
| `ShareDownloadURL(fileId, accessToken string)`                                                                                                           | `POST /nd.bizuserres.s/v1/get_share_download_url`    | `ShareDownloadURLRequest` | `ShareDownloadURLResponse` |

### 组合方法与底层方法

以下方法不是独立的 JSON 业务接口。

| 方法                                                                                                                                                              | 参数与行为                                                                                               | 返回值                                 |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | ----------------------------------- |
| `NewClient(accessToken, refreshToken, deviceID string)`                                                                                                         | 令牌可为空；`deviceID` 为空时自动生成                                                                            | `*Client`                           |
| `Close()`                                                                                                                                                       | 当前为空实现                                                                                              | 无                                   |
| `LoginSMS(phone string, getCode func() (string, error), target string)`                                                                                         | 顺序执行初始化、发送、验证、登录；`getCode` 不可为 `nil`；`target=""` 时使用 `"ANY"`                                        | `TokenResponse, error`              |
| `FSImageList(page, pageSize, orderBy, sortType int)`                                                                                                            | 调用 `FSFiles`，固定 `parentId="*"`、`fileTypes=[1]`、`resType=1`                                          | `FileListResponse, error`           |
| `FSVideoList(page, pageSize, orderBy, sortType int, play bool)`                                                                                                 | 调用 `FSFiles`，固定 `parentId="*"`、`fileTypes=[2]`、`resType=1`；`play` 对应 `needPlayRecord`               | `FileListResponse, error`           |
| `FSDocumentList(page, pageSize, orderBy, sortType int)`                                                                                                         | 调用 `FSFiles`，固定 `parentId="*"`、`fileTypes=[4]`、`resType=1`                                          | `FileListResponse, error`           |
| `FSRecycleFiles(page, pageSize, orderBy, sortType int)`                                                                                                         | 调用 `FSFiles`，固定 `parentId=""`、`dirType=4`                                                           | `FileListResponse, error`           |
| `FileUpload(path, name string, parentId any)`                                                                                                                   | `path` 为本地文件；空 `name` 使用本地文件名；申请凭证、检查秒传、必要时分片上传并查询结果                                                | `UploadInfoResponse, error`         |
| `CDNUpload(path string, tokenData UploadTokenData, contentType string, chunkSize int64)`                                                                        | `tokenData` 取自上传凭证；空 `contentType` 使用 `application/octet-stream`；`chunkSize<=0` 时用 `5*1024*1024` 字节 | 完成上传返回的 ETag（`string`）、`error`      |
| `OSSRequest(method, rawURL, bucket, objectKey, accessKey, secret, securityToken string, content []byte, contentType, contentMD5 string, sub map[string]string)` | URL、对象与凭证来自上传令牌；`content` 为原始请求体；`sub` 编码为查询参数并参与签名；`contentMD5` 为可选内容摘要                            | `*http.Response, error`；成功响应体由调用方关闭 |
| `OSSSignHeaders(method, bucket, key, accessKey, secret, securityToken, contentType, contentMD5 string, sub map[string]string)`                                  | 根据 HTTP 方法、对象、凭证、内容元数据与子资源参数生成签名头，不发送请求                                                             | `http.Header`                       |

## 请求参数详解

### 认证请求

| 请求类型                     | JSON 字段与 Go 类型                                                                                  | 默认值及说明                                                                                                        |
| ------------------------ | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `CaptchaInitRequest`     | `client_id string`、`action string`、`device_id string`、`meta CaptchaMeta`、`captcha_token string` | `client_id=clientID`；`action="POST:/v1/auth/verification"`；`device_id=Client.DeviceID`；`captcha_token=""` 时省略 |
| `CaptchaMeta`            | `phone_number string`                                                                           | 对应 `LoginSMSInit` 的 `phone`                                                                                   |
| `SMSVerificationRequest` | `phone_number string`、`target string`、`client_id string`                                        | `client_id=clientID`；验证码令牌不在 JSON 中，而在 `X-Captcha-Token` 请求头                                                  |
| `SMSVerifyRequest`       | `verification_id string`、`verification_code string`、`client_id string`                          | 前两个字段来自验证 ID 和短信验证码；`client_id=clientID`                                                                      |
| `SMSSigninRequest`       | `verification_code string`、`verification_token string`、`username string`、`client_id string`     | `client_id=clientID`；`captchaToken` 写入 `X-Captcha-Token` 请求头；成功后更新客户端令牌                                       |
| `RefreshTokenRequest`    | `client_id string`、`grant_type string`、`refresh_token string`                                   | `grant_type="refresh_token"`；方法入参为空时使用 `Client.RefreshTokenValue`，仍为空则返回错误；请求头含 `X-Action: 401`               |

### 资产、消息和记录请求

| 请求类型                       | JSON 字段与 Go 类型                                                  | 参数说明                                                                                 |
| -------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| `EmptyRequest`             | 无字段，序列化为 `{}`                                                   | 用于资产查询及清空回收站                                                                         |
| `TrafficStatisticsRequest` | `bizType int`、`groupBy int`、`startDate string`、`endDate string` | 业务类型、分组方式、开始/结束日期；日期格式 `YYYY-MM-DD`。代码注释记录 `bizType=8` 为直链流量包，`5` 为免登录流量包；未校验枚举及日期范围 |
| `InAppMsgListRequest`      | `msgType int`、`page int`、`pageSize int`                         | 消息类型、页码、每页数量；三个字段均发送                                                                 |
| `UserActionRequest`        | `pageSize int`、`cursor string`、`fileTypes []int`                | 每页数量、字符串游标、文件类型过滤；首次可传 `cursor=""`；`fileTypes=nil` 发送 `null`，空切片发送 `[]`              |
| `RestoreListRequest`       | `pageSize int`、`cursor int`、`orderBy int`、`sortType int`        | 转存列表的数量、整数游标、排序字段和排序类型；首次示例使用 `cursor=0`                                             |

### 文件请求

`FSFilesRequest` 的字段如下：

| JSON 字段          | Go 类型   | 说明                                       |
| ---------------- | ------- | ---------------------------------------- |
| `parentId`       | `any`   | 父目录 ID；`FSFiles` 方法将 `nil` 转为 `""`       |
| `page`           | `int`   | 页码，原样发送                                  |
| `pageSize`       | `int`   | 每页数量                                     |
| `orderBy`        | `int`   | 排序字段，枚举由接口约定                             |
| `sortType`       | `int`   | 排序类型，枚举由接口约定                             |
| `fileTypes`      | `[]int` | 类型过滤；当前 DTO 使用 `omitempty`，`nil` 和空切片均省略 |
| `resType`        | `*int`  | 资源类型；`nil` 省略，非空指针即使指向 `0` 也发送           |
| `dirType`        | `*int`  | 目录类型；`nil` 省略，非空指针即使指向 `0` 也发送           |
| `needPlayRecord` | `bool`  | 是否需要播放记录；仅 `true` 时发送                    |

文件类型映射见 `FileType`：图片 `1`、视频 `2`、音频 `3`、文档 `4`、压缩包 `5`、安装包 `8`、BT 种子 `9`。

| 请求类型                      | JSON 字段与 Go 类型                                         | 参数说明                                                                    |
| ------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------- |
| `SearchFilesRequest`      | `name string`、`pageSize int`                           | 搜索名称、数量；当前接口封装没有页码或游标参数                                                 |
| `CompressFileListRequest` | `fileId string`、`pageSize int`、`password string`       | 压缩文件 ID、每页数量和压缩文件密码；密码为空字符串时仍会发送 `password: ""`                         |
| `DecompressFilesRequest`  | `fileId string`、`password string`、`filePaths []string`、`toFileId string` | 压缩文件 ID、密码、待解压的文件路径列表和目标目录 ID；密码为空字符串时仍会发送 |
| `QueryDecompressStatusRequest` | `taskId string` | 解压缩任务 ID |
| `FSCreateDirRequest`      | `dirName string`、`parentId any`、`failIfNameExist bool` | 新目录名、父目录、同名时是否失败；父目录 `nil` 转 `""`，`failIfNameExist=false` 时省略           |
| `FileIDRequest`           | `fileId string`                                        | 文件详情、下载地址查询的目标文件 ID                                                     |
| `FSRenameRequest`         | `fileId string`、`newName string`                       | 目标文件 ID、新名称                                                             |
| `FileIDsRequest`          | `fileIds []any`、`parentId any`                         | 复制/移动时设置目标目录，传入 `nil` 转 `""`；删除/恢复回收站文件时不设置 `parentId`，由 `omitempty` 省略 |
| `TaskStatusRequest`       | `taskId string`                                        | 查询任务状态或上传文件信息的任务 ID                                                     |

`FileIDsRequest.ParentID` 是接口类型：复制/移动时设置的空字符串仍作为 `parentId: ""` 发送；未设置的 `nil` 接口才省略。`fileIds` 没有 `omitempty`，`nil` 序列化为 `null`。

### 云添加与上传请求

| 请求类型或编码                      | JSON/表单字段与 Go 类型                                                      | 参数说明                                                    |
| ---------------------------- | --------------------------------------------------------------------- | ------------------------------------------------------- |
| `CloudResolveURLRequest`     | `url string`                                                          | 待解析资源链接                                                 |
| `CloudCreateTaskRequest`     | `fileIndexes []int`、`url string`、`parentId any`、`newName string`    | BT 文件索引列表、待添加链接、目标目录和新名称；父目录 `nil` 转 `""` |
| `CloudTaskListRequest`       | `page int`、`pageSize int`、`status []int`                              | 页码、每页数量、任务状态；`status=nil` 时使用 `[0,1,3,4]`，显式空切片保留为 `[]` |
| Multipart `ResolveTorrent`   | 文件字段 `torrent`                                                        | `data []byte` 为种子内容，`filename string` 为上传文件名；不使用 JSON   |
| `UploadTokenRequest`         | `capacity int`、`name string`、`res UploadTokenResource`、`parentId any` | `capacity` 固定为 `30`；名称来自 `name`；父目录 `nil` 转 `""`        |
| `UploadTokenResource`        | `fileSize int64`、`md5 string`                                         | 文件字节数来自 `size`；摘要来自 `fileMD5`，为空则省略                     |
| `CheckCanFlashUploadRequest` | `taskId string`、`gcid string`、`cid string`                            | `taskId` 来自入参；`gcid`、`cid` 由本地 `path` 文件计算；`path` 本身不发送 |

`FileUpload` 对小于 1 MiB 的文件自动计算 MD5，并将其传给 `UploadToken`。成功秒传或完成 OSS 上传后，调用 `UploadInfo`，最多查询 4 次、重试间隔 5 秒；上传中返回 `code=147` 或 `msg="文件上传中"` 时继续等待。

### 分享请求

| 请求类型                      | JSON 字段与 Go 类型                                                                                                                                                         | 参数说明                                                           |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `ShareCreateRequest`      | `fileIds []any`、`title string`、`validateDuration int`、`shareType int`、`code string`、`autoFillCode bool`、`trafficLimit string`、`maxRestoreCount int`、`downloadType int` | `ShareCreate` 只接收文件 ID 列表和标题，其余使用下列固定值；所有字段均发送                 |
| `ShareUpdateRequest`      | `id string`、`title string`、`validateDuration int`、`shareType int`、`code string`、`autoFillCode bool`、`trafficLimit string`、`maxRestoreCount int`、`downloadType int`     | 分享 ID、标题、有效时长、分享类型、提取码、是否自动填码、流量限制、最大转存次数、下载类型；全部由方法参数提供，零值也发送 |
| `ShareListRequest`        | `page int`、`pageSize int`、`orderType int`、`sortType int`                                                                                                               | 页码、每页数量、排序字段及类型；方法参数 `orderBy` 写入的是 `orderType`，不是 `orderBy`   |
| `ShareDeleteRequest`      | `ids []any`                                                                                                                                                            | 要删除的分享 ID 列表                                                   |
| `ShareRestoreRequest`     | `accessToken string`、`fileIds []any`、`parentId string`                                                                                                                 | 分享访问令牌、转存文件 ID 列表、目标父目录；空父目录仍发送                                |
| `ShareIDRequest`          | `shareId string`                                                                                                                                                       | 分享摘要查询                                                         |
| `ShareAccessTokenRequest` | `shareId string`、`code string`                                                                                                                                         | 分享 ID 和提取码；空 `code` 仍发送                                        |
| `ShareFilesListRequest`   | `accessToken string`、`parentId string`、`page int`、`pageSize int`、`orderBy int`、`sortType int`                                                                          | 分享访问令牌、父目录、页码、每页数量、排序字段及类型                                     |
| `ShareFilesSizeRequest`   | `accessToken string`、`fileIds []any`、`download bool`                                                                                                                   | 分享访问令牌、文件 ID 列表、下载标志；`false` 仍发送                               |
| `ShareDownloadURLRequest` | `fileId string`、`accessToken string`                                                                                                                                   | 分享中的文件 ID 及分享访问令牌                                              |

`ShareCreate` 除 `fileIds`、`title` 外的固定请求值：

```json
{
  "validateDuration": 0,
  "shareType": 1,
  "code": "",
  "autoFillCode": true,
  "trafficLimit": "0",
  "maxRestoreCount": 0,
  "downloadType": 1
}
```

这些值是当前客户端默认值；未确认的服务端枚举含义、时间或流量单位不在此推断。

<a id="api-responses"></a>

## 响应结构详解

### 通用响应与错误

大多数业务接口返回 `APIResponse[T]`：

| JSON 字段 | Go 字段/类型     | 说明                          |
| ------- | ------------ | --------------------------- |
| `data`  | `Data T`     | 业务数据，具体类型见下表                |
| `msg`   | `Msg string` | 消息，可在响应中缺省，解码为 `""`         |
| `code`  | `Code int`   | 可缺省，解码为 `0`；不应仅凭缺省值推断业务一定成功 |

下文未特别标注的字段都位于 `data` 内。表格分组只为方便阅读，不改变 JSON 层级。可选字段缺省时通常得到 Go 零值；切片可为 `nil`，指针可为 `nil`，调用方应按实际结果处理。

公共请求层返回网络、非 2xx HTTP 状态和 JSON 编解码错误，但**不会统一把业务** **`code != 0`** **转为** **`error`**；`FileUpload` 等组合流程另有业务判断。`APIError` 结构包含顶层 `error string`、`error_code int`、`error_description string`、`details []json.RawMessage`，目前公共请求层不会自动把它转换为类型化错误。

`GenericResponse` 是 `APIResponse[json.RawMessage]`，用于保留尚未建模的数据；本接口列表中的业务方法使用各自明确的响应类型。

### 认证顶层响应

| 返回类型                      | 顶层 JSON 字段与 Go 类型                                                                                             |
| ------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `CaptchaInitResponse`     | `captcha_token string`、`url string`（可选）                                                                       |
| `SMSVerificationResponse` | `verification_id string`                                                                                      |
| `SMSVerifyResponse`       | `verification_token string`                                                                                   |
| `TokenResponse`           | `access_token string`、`expires_in int`、`refresh_token string`、`scope string`、`sub string`、`token_type string` |
| `UserInfoResponse`        | `sub string`、`name string`、`phone_number string`、`created_at time.Time`、`password_updated_at time.Time`       |

`TokenResponse.expires_in` 按秒用于计算 `Client.TokenExpiresAt`；`sub` 对应 Go 字段 `Subject`。`time.Time` 字段接收 Go 标准时间解码器支持的 RFC3339 格式字符串。

### 资产与消息

`AssetsResponse = APIResponse[AssetsData]`：

| 字段                                               | Go 类型   | 说明                        |
| ------------------------------------------------ | ------- | ------------------------- |
| `totalSpaceSize`、`usedSpaceSize`                 | `int64` | 总容量、已用容量                  |
| `totalDirectLinkTraffic`、`freeDirectLinkTraffic` | `int64` | 直链总流量、可用流量                |
| `totalShareGuestTraffic`、`freeShareGuestTraffic` | `int64` | 分享访客总流量、可用流量              |
| `vipStatus`、`svipStatus`                         | `int`   | 会员状态                      |
| `vipLeftTime`、`vipExpireTime`、`systemTime`       | `int64` | 会员剩余时间、到期时间、系统时间，保留接口原始数值 |

`TrafficStatisticsResponse = APIResponse[TrafficStatisticsData]`：

| 字段                                           | Go 类型                     | 说明                     |
| -------------------------------------------- | ------------------------- | ---------------------- |
| `totalTraffic`、`totalRemained`               | `int64`                   | 总流量、剩余流量               |
| `rewardTotalTraffic`、`rewardRemainedTraffic` | `int64`                   | 奖励总流量、奖励剩余流量           |
| `statistics`                                 | `[]TrafficStatisticsItem` | 分组统计列表                 |
| `statistics[].key`                           | `string`                  | 分组键，日期示例为 `2026-08-16` |

当前 `TrafficStatisticsItem` 仅建模 `key`，没有推断或补充未出现在样例中的统计值字段。

`InAppMsgListResponse = APIResponse[InAppMsgListData]`：

| 字段                           | Go 类型               | 说明            |
| ---------------------------- | ------------------- | ------------- |
| `total`                      | `int`               | 总条数           |
| `list`                       | `[]InAppMsgItem`    | 消息列表          |
| `list[].id`                  | `int`               | 消息 ID         |
| `list[].title`               | `string`            | 标题            |
| `list[].ctime`               | `int64`             | 创建时间          |
| `list[].status`              | `int`               | 消息状态          |
| `list[].contents`            | `[]InAppMsgContent` | 内容段列表         |
| `list[].contents[].content`  | `string`            | 文本内容          |
| `list[].contents[].fontSize` | `int`               | 可选字号；缺省时为 `0` |

### 文件列表和详情

| 返回类型                       | Data 类型                | Data 字段                                                                                                                                        |
| -------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `FileListResponse`         | `FileListData`         | `total int`、`list []FileItem`；搜索接口也使用此结构                                                                                                       |
| `CompressFileListResponse` | `CompressFileListData` | `total int`、`pageSize int`、`list []CompressFileItem`                                                                                           |
| `FileTaskResponse`         | `FileTaskData`         | `taskId string`；解压缩接口返回异步任务 ID                                                                                                        |
| `RestoreListResponse`      | `RestoreListData`      | `total int`、`list []FileItem`、`cursor int`、`hasMore bool`                                                                                      |
| `ShareFilesListResponse`   | `ShareFilesListData`   | `total int`、`list []FileItem`、`cursor int`；没有 `hasMore` 字段                                                                                     |
| `FSCreateDirResponse`      | `FSCreateDirData`      | `fileId string`、`fileName string`、`parentId string`、`depth int`、`dirType int`、`resType int`、`fullParentIds string`、`ctime int64`、`utime int64` |
| `FileDetailResponse`       | `FileDetailData`       | `fileInfo FileDetailInfo`、`location string`、`picInfo PictureInfo`                                                                              |
| `UploadInfoResponse`       | `*FileDetailInfo`      | 上传中数据可能为空；访问前检查 `Data != nil`                                                                                                                  |
| `FileTaskResponse`         | `FileTaskData`         | `taskId string`；用于删除、复制、移动、转存产生的任务                                                                                                             |
| `TaskStatusResponse`       | `TaskStatusData`       | `status int`、`progress int`；未在 SDK 中定义完整状态枚举                                                                                                  |
| `EmptyResponse`            | `*struct{}`            | `data: null` 对应 `Data == nil`                                                                                                                  |
| `DownloadURLResponse`      | `DownloadURLData`      | `requestId string`、`signedURL string`、`speedupSignature string`、`urlDuration int`                                                              |

`FileItem` 的完整字段如下。目录条目可能没有文件大小、摘要、缩略图等文件专属字段。

| JSON 字段                 | Go 类型    | 说明                                                 |
| ----------------------- | -------- | -------------------------------------------------- |
| `fileId`、`fileName`     | `string` | 文件/目录 ID 与名称                                       |
| `parentId`、`parentName` | `string` | 父目录 ID 与名称，可缺省；搜索响应包含 `parentName`                 |
| `fileSize`              | `int64`  | 文件大小，可缺省；支持超过 32 位整数的值                             |
| `gcid`、`md5`            | `string` | 文件摘要，可缺省                                           |
| `depth`                 | `int`    | 目录深度                                               |
| `mineType`              | `string` | MIME 类型，可缺省；JSON 拼写是 `mineType`，Go 字段名为 `MimeType` |
| `fileType`              | `int`    | 文件类型，可缺省                                           |
| `dirType`、`resType`     | `int`    | 目录类型、资源类型                                          |
| `ext`                   | `string` | 扩展名，可缺省                                            |
| `fullParentIds`         | `string` | 完整父目录 ID 路径，可缺省                                    |
| `ctime`、`utime`         | `int64`  | 创建、更新时间                                            |
| `thumbnail`             | `string` | 缩略图地址，可缺省                                          |
| `auditStatus`           | `int`    | 审核状态                                               |
| `leftTime`              | `int64`  | 剩余时间，可缺省                                           |

`FileDetailInfo` 包含上述 `FileItem` 字段中的全部字段，**除** **`parentName`、`leftTime`** **外**。`PictureInfo` 包含可选的 `previewUrl string`。

### 用户操作记录

`UserActionResponse = APIResponse[UserActionData]`：

| 字段                                | Go 类型                | 说明                                           |
| --------------------------------- | -------------------- | -------------------------------------------- |
| `total`                           | `int`                | 记录总数                                         |
| `cursor`                          | `string`             | 下一次查询使用的游标                                   |
| `hasMore`                         | `bool`               | 是否还有后续记录                                     |
| `list`                            | `[]UserActionItem`   | 操作分组列表                                       |
| `list[].collectionId`、`list[].id` | `string`             | 分组标识与记录 ID                                   |
| `list[].actionType`               | `int`                | 当前代码注释记录：`2` 手机上传、`3` 网页上传、`5` 手机播放、`6` 网页播放 |
| `list[].ctime`                    | `int64`              | 操作时间                                         |
| `list[].totalCount`               | `int`                | 分组包含的条数                                      |
| `list[].actionDetails`            | `[]UserActionDetail` | 操作文件明细                                       |

`UserActionDetail`：

| 字段                                 | Go 类型    | 说明                 |
| ---------------------------------- | -------- | ------------------ |
| `fileId`、`fileName`、`thumbnail`    | `string` | 文件 ID、名称、缩略图       |
| `fileSize`、`utime`                 | `int64`  | 文件大小、更新时间          |
| `parentId`、`parentName`            | `string` | 父目录 ID、名称          |
| `fileType`、`resType`、`auditStatus` | `int`    | 文件类型、资源类型、审核状态     |
| `playMilliSeconds`、`duration`      | `int64`  | 播放进度及媒体时长，样例使用毫秒数值 |
| `ext`                              | `string` | 文件扩展名              |

### 云添加响应

| 返回类型                          | Data 类型及字段                                                                                           |
| ----------------------------- | ---------------------------------------------------------------------------------------------------- |
| `CloudResolveURLResponse`     | `CloudResolveURLData`：`resType int`、`btResInfo BTResourceInfo`、`url string`                          |
| `CloudResolveTorrentResponse` | `CloudResolveTorrentData`：`resType int`、`btResInfo BTResourceInfo`                                   |
| `CloudCreateTaskResponse`     | `CloudCreateTaskData`：`taskId string`、`url string`                                                   |
| `CloudTaskListResponse`       | `CloudTaskListData`：`cursor string`、`list []CloudTask`、`statusCounts []CloudStatusCount`、`total int` |

嵌套类型：

| 类型                 | 字段与 Go 类型                                                                                                                                                                        |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BTResourceInfo`   | `infoHash string`、`fileName string`、`fileSize int64`、`subfilesNum int`、`subfiles []CloudSubfile`；可选 `createTime int64`、`excludeIndices []int`                                    |
| `CloudSubfile`     | `fileName string`；可选 `fileIndex *int`、`fileSize int64`、`isDir bool`、`subfiles []CloudSubfile`（递归子文件结构）                                                                           |
| `CloudTask`        | `createTime int64`、`fileName string`、`isDir bool`、`parentDirType int`、`res string`、`resType int`、`status int`、`taskId string`、`totalSize int64`；可选 `errCode int`、`errMsg string` |
| `CloudStatusCount` | `status int`；可选 `count int`                                                                                                                                                      |

### 分享响应

| 返回类型                       | 响应结构                                                                                                                      |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `ShareCreateResponse`      | `Data ShareCreateData`：`code string`、`createTime string`、`shareId string`、`shareUrl string`                               |
| `ShareListResponse`        | `Data ShareListData`：`list []ShareItem`、`total int`                                                                       |
| `ShareSummaryResponse`     | `Data ShareSummaryData`，见下表                                                                                               |
| `ShareAccessTokenResponse` | 非通用包装：顶层 `msg string`、`data ShareAccessTokenData`、可选 `access_token string`；`ShareAccessTokenData` 包含 `accessToken string` |
| `ShareFilesSizeResponse`   | `Data ShareFilesSizeData`：`totalSize int64`、`fileSizeMap map[string]int64`（键为文件 ID）                                       |
| `ShareDownloadURLResponse` | `Data ShareDownloadURLData`：`downloadUrl string`                                                                          |

`ShareAccessTokenResponse` 同时保留 `result.Data.AccessToken` 与 `result.AccessToken` 两种位置，不会自动合并；调用方应按实际响应选用。

`ShareItem`：

| 字段                              | Go 类型    | 说明                                                 |
| ------------------------------- | -------- | -------------------------------------------------- |
| `id`、`shareId`、`title`、`shareUrl` | `string` | 记录 ID、分享 ID、标题与分享地址 |
| `code`、`thumbnail` | `string` | 提取码与缩略图地址，可缺省 |
| `createTime`、`leftTime` | `int64` | 创建时间与剩余时间，保留 `leftTime: -1`；注意与 `ShareCreateData.createTime` 的字符串类型不同 |
| `shareStatus`、`resType`、`downloadType` | `int` | 分享状态、资源类型与下载类型 |
| `fileType`、`shareType` | `int` | 文件类型与分享类型，可缺省 |
| `autoFillCode`、`hasAuditReject`、`isMultiFileShare` | `bool` | 自动填充提取码、存在审核拒绝、多文件分享标记，可缺省 |
| `accessToken` | `string` | 保留的兼容字段：访问令牌，可缺省 |
| `validateDuration` | `int` | 保留的兼容字段：有效时长，可缺省 |

`ShareListData` 的 `total` 和 `list` 不省略；条目中的可选字段缺省时使用对应 Go 类型的零值。

`ShareSummaryData`：

| 字段                                     | Go 类型    | 说明                |
| -------------------------------------- | -------- | ----------------- |
| `nickName`、`userId`、`title`、`shareId`  | `string` | 昵称、用户 ID、标题、分享 ID |
| `ctime`、`leftTime`                     | `int64`  | 创建时间、剩余时间         |
| `needCode`                             | `bool`   | 是否需要提取码           |
| `shareStatus`、`vipStatus`、`svipStatus` | `int`    | 分享与会员状态           |
| `totalFileNum`                         | `int`    | 文件总数              |
| `totalFileSize`                        | `int64`  | 文件总大小             |

### 上传响应

`UploadTokenResponse = APIResponse[UploadTokenData]`：

| 字段                                       | Go 类型               | 说明            |
| ---------------------------------------- | ------------------- | ------------- |
| `taskId`                                 | `string`            | 上传任务 ID       |
| `creds`                                  | `UploadCredentials` | 临时上传凭证        |
| `fullEndPoint`、`bucketName`、`objectPath` | `string`            | 完整端点、存储桶、对象路径 |
| `gcid`                                   | `string`            | 可选文件摘要        |
| `provider`                               | `int`               | 可选存储提供方       |
| `endPoint`、`callbackVar`、`region`        | `string`            | 可选端点、回调参数、地域  |

`UploadCredentials` 包含 `accessKeyID string`、`secretAccessKey string`、`sessionToken string` 和可选 `expiration *time.Time`。这些值用于 OSS 签名，不应写入公开日志或文档样例。

`FlashUploadResponse = APIResponse[FlashUploadData]`，其中 `canFlashUpload bool` 表示能否秒传。`UploadInfoResponse` 的具体文件信息见“文件列表和详情”；返回 `data: null` 时 `Data` 为 `nil`。

<a id="api-examples"></a>

## 新增接口请求与响应示例

以下 JSON 为字段说明用的缩略样例，非实时接口结果；列表省略了重复条目，`total` 不一定等于示例中的 `list` 长度。Go 调用示例中的错误处理沿用上文。

### 流量统计

```go
result, err := client.GetTrafficStatistics(8, 0, "2026-08-16", "2026-09-14")
```

请求：

```json
{"bizType":8,"groupBy":0,"startDate":"2026-08-16","endDate":"2026-09-14"}
```

响应（允许省略 `msg` 和 `code`）：

```json
{
  "data": {
    "totalTraffic": 21474836480,
    "totalRemained": 21474836480,
    "rewardTotalTraffic": 21474836480,
    "rewardRemainedTraffic": 21474836480,
    "statistics": [{"key": "2026-08-16"}]
  }
}
```

### 站内消息

```go
result, err := client.GetInAppMsgList(1, 1, 10)
```

请求（当前 SDK 会额外发送 `page`）：

```json
{"msgType":1,"page":1,"pageSize":10}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "total": 2,
    "list": [
      {"id":419129,"title":"赠送会员到账通知","contents":[{"content":"赠送会员时长已到账，请到会员中心查看。","fontSize":14}],"ctime":1780922238,"status":1},
      {"id":38821,"title":"欢迎使用光鸭云盘","contents":[{"content":"欢迎使用光鸭云盘。"}],"ctime":1776670058,"status":1}
    ]
  }
}
```

### 解压缩文件

```go
result, err := client.DecompressFiles(
	"1946490208080760901",
	"",
	[]string{"README.md"},
	"1902206196381814882",
)
```

请求：

```json
{
  "fileId": "1946490208080760901",
  "password": "",
  "filePaths": ["README.md"],
  "toFileId": "1902206196381814882"
}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "taskId": "ZB61033440"
  }
}
```

### 查询解压缩任务状态

```go
result, err := client.QueryDecompressStatus("ZB61039107")
```

请求：

```json
{"taskId":"ZB61039107"}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "status": 1,
    "progress": 99
  }
}
```

### 用户操作记录

```go
result, err := client.GetUserAction(2, "", []int{2})
```

请求：

```json
{"pageSize":2,"cursor":"","fileTypes":[2]}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "total": 48,
    "list": [{
      "collectionId": "2026090516000602",
      "actionType": 6,
      "ctime": 1788598663,
      "totalCount": 1,
      "actionDetails": [{
        "fileId":"1942451290842243086","fileName":"example.mkv",
        "thumbnail":"https://example.com/thumbnail.jpg","fileSize":4488982900,
        "utime":1788598854,"parentId":"1942451286965903415","parentName":"电影",
        "fileType":2,"resType":1,"playMilliSeconds":131236,"duration":6401185,
        "ext":".mkv","auditStatus":2
      }],
      "id": "1943259055998869558"
    }],
    "cursor": "1943259055998869558",
    "hasMore": true
  }
}
```

检查 `result.Data.HasMore` 后，将 `result.Data.Cursor` 原样传入下一次查询，不要转成整数。

### 转存列表

```go
result, err := client.GetRestoreList(4, 0, 2, 1)
```

请求：

```json
{"pageSize":4,"cursor":0,"orderBy":2,"sortType":1}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "total": 80,
    "list": [{
      "fileId":"1899691949784735777","fileName":"玻璃樽",
      "parentId":"1899684988037038116","depth":3,"dirType":1,"resType":2,
      "fullParentIds":"1893897889156169818/1899684988037038116/",
      "ctime":1778211455,"utime":1778211552,"auditStatus":4
    }],
    "cursor": 4,
    "hasMore": true
  }
}
```

下一次调用使用返回的整数 `cursor`，不要自行按页码替换。

### 搜索文件

```go
result, err := client.SearchFiles("死侍", 20)
```

请求：

```json
{"name":"死侍","pageSize":20}
```

响应（文件与目录共用 `FileItem`）：

```json
{
  "msg": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "fileId":"1903118122813124674","fileName":"死侍 (2016)",
        "parentId":"1894360528676503650","depth":3,"dirType":1,"resType":2,
        "fullParentIds":"1893897889156169818/1894360528676503650/",
        "ctime":1779028319,"utime":1779028325,"auditStatus":4,"parentName":"外语电影"
      },
      {
        "fileId":"1903118023387172879","fileName":"死侍 (2016).mkv","fileSize":3683248152,
        "gcid":"AE2797D9FC670B62C5BC455DE0D0C725D4AB3F74",
        "parentId":"1903118122813124674","depth":4,"mineType":"video/x-matroska",
        "fileType":2,"dirType":1,"resType":1,"ext":".mkv",
        "fullParentIds":"1893897889156169818/1894360528676503650/1903118122813124674/",
        "ctime":1779028295,"utime":1779028324,"md5":"71AEBFA61BAD59D441F73C4C58E6E756",
        "thumbnail":"https://example.com/thumbnail.jpg","auditStatus":4,"parentName":"死侍 (2016)"
      }
    ]
  }
}
```

### 压缩文件列表

```go
result, err := client.GetCompressFileList("1946490208080760901", 100, "")
```

请求：

```json
{"fileId":"1946490208080760901","pageSize":100,"password":""}
```

响应：

```json
{
  "msg": "success",
  "data": {
    "total": 2,
    "pageSize": 100,
    "list": [
      {
        "name": ".cache",
        "isDir": true,
        "fullPath": ".cache"
      },
      {
        "fileIndex": 2306,
        "name": "user_info_test.go",
        "fileSize": 826,
        "mimeType": "application/octet-stream",
        "fullPath": "user_info_test.go",
        "fileType": 10
      }
    ]
  }
}
```

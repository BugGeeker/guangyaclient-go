package guangyaclient

import (
	"encoding/json"
	"time"
)

// APIResponse is the common response envelope used by most api.guangyapan.com
// endpoints.
type APIResponse[T any] struct {
	Data T      `json:"data"`
	Msg  string `json:"msg"`
	// Code is optional; successful responses can omit it.
	Code int `json:"code,omitempty"`
}

type APIError struct {
	Error            string            `json:"error"`
	ErrorCode        int               `json:"error_code"`
	ErrorDescription string            `json:"error_description"`
	Details          []json.RawMessage `json:"details,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Subject      string `json:"sub"`
	TokenType    string `json:"token_type"`
}

type UserInfoResponse struct {
	Subject           string    `json:"sub"`
	Name              string    `json:"name"`
	PhoneNumber       string    `json:"phone_number"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordUpdatedAt time.Time `json:"password_updated_at"`
}

type AssetsData struct {
	TotalSpaceSize         int64 `json:"totalSpaceSize"`
	UsedSpaceSize          int64 `json:"usedSpaceSize"`
	TotalDirectLinkTraffic int64 `json:"totalDirectLinkTraffic"`
	FreeDirectLinkTraffic  int64 `json:"freeDirectLinkTraffic"`
	VIPStatus              int   `json:"vipStatus"`
	VIPLeftTime            int64 `json:"vipLeftTime"`
	SVIPStatus             int   `json:"svipStatus"`
	TotalShareGuestTraffic int64 `json:"totalShareGuestTraffic"`
	FreeShareGuestTraffic  int64 `json:"freeShareGuestTraffic"`
	VIPExpireTime          int64 `json:"vipExpireTime"`
	SystemTime             int64 `json:"systemTime"`
}

type AssetsResponse = APIResponse[AssetsData]

type TrafficStatisticsItem struct {
	Key string `json:"key"`
}

type TrafficStatisticsData struct {
	TotalTraffic          int64                   `json:"totalTraffic"`
	TotalRemained         int64                   `json:"totalRemained"`
	RewardTotalTraffic    int64                   `json:"rewardTotalTraffic"`
	RewardRemainedTraffic int64                   `json:"rewardRemainedTraffic"`
	Statistics            []TrafficStatisticsItem `json:"statistics"`
}

type TrafficStatisticsResponse = APIResponse[TrafficStatisticsData]

type InAppMsgContent struct {
	Content  string `json:"content"`
	FontSize int    `json:"fontSize,omitempty"`
}

type InAppMsgItem struct {
	ID       int               `json:"id"`
	Title    string            `json:"title"`
	Contents []InAppMsgContent `json:"contents"`
	CTime    int64             `json:"ctime"`
	Status   int               `json:"status"`
}

type InAppMsgListData struct {
	Total int            `json:"total"`
	List  []InAppMsgItem `json:"list"`
}

type InAppMsgListResponse = APIResponse[InAppMsgListData]

type UserActionDetail struct {
	FileID           string `json:"fileId"`
	FileName         string `json:"fileName"`
	Thumbnail        string `json:"thumbnail"`
	FileSize         int64  `json:"fileSize"`
	UTime            int64  `json:"utime"`
	ParentID         string `json:"parentId"`
	ParentName       string `json:"parentName"`
	FileType         int    `json:"fileType"`
	ResType          int    `json:"resType"`
	PlayMilliSeconds int64  `json:"playMilliSeconds"`
	Duration         int64  `json:"duration"`
	Ext              string `json:"ext"`
	AuditStatus      int    `json:"auditStatus"`
}

type UserActionItem struct {
	CollectionID  string             `json:"collectionId"`
	ActionType    int                `json:"actionType"` // 2：手机上传 3：网页上传 5：手机播放 6：网页播放
	CTime         int64              `json:"ctime"`
	TotalCount    int                `json:"totalCount"`
	ActionDetails []UserActionDetail `json:"actionDetails"`
	ID            string             `json:"id"`
}

type UserActionData struct {
	Total   int              `json:"total"`
	List    []UserActionItem `json:"list"`
	Cursor  string           `json:"cursor"`
	HasMore bool             `json:"hasMore"`
}

type UserActionResponse = APIResponse[UserActionData]

type CaptchaInitResponse struct {
	CaptchaToken string `json:"captcha_token"`
	URL          string `json:"url,omitempty"`
}

type SMSVerificationResponse struct {
	VerificationID string `json:"verification_id"`
}

type SMSVerifyResponse struct {
	VerificationToken string `json:"verification_token"`
}

// GenericResponse preserves the data of endpoints without a verified data schema.
type GenericResponse = APIResponse[json.RawMessage]

type FileIDRequest struct {
	FileID string `json:"fileId"`
}

type EmptyRequest struct{}

type FSFilesRequest struct {
	ParentID       any   `json:"parentId"`
	Page           int   `json:"page"`
	PageSize       int   `json:"pageSize"`
	OrderBy        int   `json:"orderBy"`
	SortType       int   `json:"sortType"`
	FileTypes      []int `json:"fileTypes,omitempty"`
	ResType        *int  `json:"resType,omitempty"`
	DirType        *int  `json:"dirType,omitempty"`
	NeedPlayRecord bool  `json:"needPlayRecord,omitempty"`
}

type CloudTaskListRequest struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Status   []int `json:"status"`
}

type UploadTokenResource struct {
	FileSize int64  `json:"fileSize"`
	MD5      string `json:"md5,omitempty"`
}

type UploadTokenRequest struct {
	Capacity int                 `json:"capacity"`
	Name     string              `json:"name"`
	Resource UploadTokenResource `json:"res"`
	ParentID any                 `json:"parentId"`
}

type CheckCanFlashUploadRequest struct {
	TaskID string `json:"taskId"`
	GCID   string `json:"gcid"`
	CID    string `json:"cid"`
}

type CaptchaMeta struct {
	PhoneNumber string `json:"phone_number"`
}

type CaptchaInitRequest struct {
	ClientID     string      `json:"client_id"`
	Action       string      `json:"action"`
	DeviceID     string      `json:"device_id"`
	Meta         CaptchaMeta `json:"meta"`
	CaptchaToken string      `json:"captcha_token,omitempty"`
}

type SMSVerificationRequest struct {
	PhoneNumber string `json:"phone_number"`
	Target      string `json:"target"`
	ClientID    string `json:"client_id"`
}

type SMSVerifyRequest struct {
	VerificationID   string `json:"verification_id"`
	VerificationCode string `json:"verification_code"`
	ClientID         string `json:"client_id"`
}

type SMSSigninRequest struct {
	VerificationCode  string `json:"verification_code"`
	VerificationToken string `json:"verification_token"`
	Username          string `json:"username"`
	ClientID          string `json:"client_id"`
}

type RefreshTokenRequest struct {
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

type DeviceTokenRequest struct {
	GrantType  string `json:"grant_type"`
	DeviceCode string `json:"device_code"`
	ClientID   string `json:"client_id"`
}

type DeviceCodeRequest struct {
	Scope    string `json:"scope"`
	ClientID string `json:"client_id"`
}

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	VerificationURL         string `json:"verification_url"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

type TrafficStatisticsRequest struct {
	BizType   int    `json:"bizType"`
	GroupBy   int    `json:"groupBy"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type InAppMsgListRequest struct {
	MsgType  int `json:"msgType"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type UserActionRequest struct {
	PageSize  int    `json:"pageSize"`
	Cursor    string `json:"cursor"`
	FileTypes []int  `json:"fileTypes"`
}

type RestoreListRequest struct {
	PageSize int `json:"pageSize"`
	Cursor   int `json:"cursor"`
	OrderBy  int `json:"orderBy"`
	SortType int `json:"sortType"`
}

type CloudResolveURLRequest struct {
	URL string `json:"url"`
}

type CloudCreateTaskRequest struct {
	FileIndexes []int  `json:"fileIndexes"`
	URL         string `json:"url"`
	ParentID    any    `json:"parentId"`
	NewName     string `json:"newName"`
}

type FSRenameRequest struct {
	FileID  string `json:"fileId"`
	NewName string `json:"newName"`
}

type ShareUpdateRequest struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	ValidateDuration int    `json:"validateDuration"`
	ShareType        int    `json:"shareType"`
	Code             string `json:"code"`
	AutoFillCode     bool   `json:"autoFillCode"`
	TrafficLimit     string `json:"trafficLimit"`
	MaxRestoreCount  int    `json:"maxRestoreCount"`
	DownloadType     int    `json:"downloadType"`
}

type ShareListRequest struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	OrderType int `json:"orderType"`
	SortType  int `json:"sortType"`
}

type ShareDeleteRequest struct {
	IDs []any `json:"ids"`
}

type ShareRestoreRequest struct {
	AccessToken string `json:"accessToken"`
	FileIDs     []any  `json:"fileIds"`
	ParentID    string `json:"parentId"`
}

type TaskStatusRequest struct {
	TaskID string `json:"taskId"`
}

type ShareDownloadURLRequest struct {
	FileID      string `json:"fileId"`
	AccessToken string `json:"accessToken"`
}

type ShareFilesSizeRequest struct {
	AccessToken string `json:"accessToken"`
	FileIDs     []any  `json:"fileIds"`
	Download    bool   `json:"download"`
}

type ShareIDRequest struct {
	ShareID string `json:"shareId"`
}

type ShareAccessTokenRequest struct {
	ShareID string `json:"shareId"`
	Code    string `json:"code"`
}

type ShareFilesListRequest struct {
	AccessToken string `json:"accessToken"`
	ParentID    string `json:"parentId"`
	Page        int    `json:"page"`
	PageSize    int    `json:"pageSize"`
	OrderBy     int    `json:"orderBy"`
	SortType    int    `json:"sortType"`
}

type SearchFilesRequest struct {
	Name     string `json:"name"`
	PageSize int    `json:"pageSize"`
}

type CompressFileListRequest struct {
	FileID   string `json:"fileId"`
	PageSize int    `json:"pageSize"`
	Password string `json:"password"`
}

type DecompressFilesRequest struct {
	FileID    string   `json:"fileId"`
	Password  string   `json:"password"`
	FilePaths []string `json:"filePaths"`
	ToFileID  string   `json:"toFileId"`
}

type QueryDecompressStatusRequest struct {
	TaskID string `json:"taskId"`
}

type FSCreateDirRequest struct {
	DirName         string `json:"dirName"`
	ParentID        any    `json:"parentId"`
	FailIfNameExist bool   `json:"failIfNameExist,omitempty"`
}

type FileIDsRequest struct {
	FileIDs  []any `json:"fileIds"`
	ParentID any   `json:"parentId,omitempty"`
}

type ShareCreateRequest struct {
	FileIDs          []any  `json:"fileIds"`
	Title            string `json:"title"`
	ValidateDuration int    `json:"validateDuration"`
	ShareType        int    `json:"shareType"`
	Code             string `json:"code"`
	AutoFillCode     bool   `json:"autoFillCode"`
	TrafficLimit     string `json:"trafficLimit"`
	MaxRestoreCount  int    `json:"maxRestoreCount"`
	DownloadType     int    `json:"downloadType"`
}

// ShareAccessTokenResponse accepts both token locations handled by the client.
type ShareAccessTokenResponse struct {
	Data        ShareAccessTokenData `json:"data"`
	Msg         string               `json:"msg"`
	AccessToken string               `json:"access_token,omitempty"`
}

type ShareAccessTokenData struct {
	AccessToken string `json:"accessToken"`
}

type ShareSummaryData struct {
	NickName      string `json:"nickName"`
	CTime         int64  `json:"ctime"`
	LeftTime      int64  `json:"leftTime"`
	NeedCode      bool   `json:"needCode"`
	UserID        string `json:"userId"`
	Title         string `json:"title"`
	ShareStatus   int    `json:"shareStatus"`
	VIPStatus     int    `json:"vipStatus"`
	SVIPStatus    int    `json:"svipStatus"`
	TotalFileNum  int    `json:"totalFileNum"`
	ShareID       string `json:"shareId"`
	TotalFileSize int64  `json:"totalFileSize"`
}

type ShareSummaryResponse = APIResponse[ShareSummaryData]

type ShareFilesListData struct {
	Total  int        `json:"total"`
	List   []FileItem `json:"list"`
	Cursor int        `json:"cursor"`
}

type ShareFilesListResponse = APIResponse[ShareFilesListData]

type FileItem struct {
	AuditStatus   int    `json:"auditStatus"`
	CTime         int64  `json:"ctime"`
	Depth         int    `json:"depth"`
	DirType       int    `json:"dirType"`
	Ext           string `json:"ext,omitempty"`
	FileID        string `json:"fileId"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize,omitempty"`
	FileType      int    `json:"fileType,omitempty"`
	FullParentIDs string `json:"fullParentIds,omitempty"`
	GCID          string `json:"gcid,omitempty"`
	MD5           string `json:"md5,omitempty"`
	MimeType      string `json:"mineType,omitempty"`
	ParentID      string `json:"parentId,omitempty"`
	ParentName    string `json:"parentName,omitempty"`
	ResType       int    `json:"resType"`
	Thumbnail     string `json:"thumbnail,omitempty"`
	LeftTime      int64  `json:"leftTime,omitempty"`
	UTime         int64  `json:"utime"`
}

type FileListData struct {
	List  []FileItem `json:"list"`
	Total int        `json:"total"`
}

type FileListResponse = APIResponse[FileListData]

type CompressFileItem struct {
	FileIndex int    `json:"fileIndex,omitempty"`
	Name      string `json:"name"`
	IsDir     bool   `json:"isDir,omitempty"`
	FileSize  int64  `json:"fileSize,omitempty"`
	MimeType  string `json:"mimeType,omitempty"`
	FullPath  string `json:"fullPath"`
	FileType  int    `json:"fileType,omitempty"`
}

type CompressFileListData struct {
	Total    int                `json:"total"`
	PageSize int                `json:"pageSize"`
	List     []CompressFileItem `json:"list"`
}

type CompressFileListResponse = APIResponse[CompressFileListData]

type RestoreListData struct {
	Total   int        `json:"total"`
	List    []FileItem `json:"list"`
	Cursor  int        `json:"cursor"`
	HasMore bool       `json:"hasMore"`
}

type RestoreListResponse = APIResponse[RestoreListData]

type FSCreateDirData struct {
	FileID        string `json:"fileId"`
	FileName      string `json:"fileName"`
	ParentID      string `json:"parentId"`
	Depth         int    `json:"depth"`
	DirType       int    `json:"dirType"`
	ResType       int    `json:"resType"`
	FullParentIDs string `json:"fullParentIds"`
	CTime         int64  `json:"ctime"`
	UTime         int64  `json:"utime"`
}

type FSCreateDirResponse = APIResponse[FSCreateDirData]

type FileTaskData struct {
	TaskID string `json:"taskId"`
}

// FileTaskResponse is returned by file copy, move, delete, and share restore operations.
type FileTaskResponse = APIResponse[FileTaskData]

type TaskStatusData struct {
	Status   int `json:"status"`   // 任务状态 1: 进行中 2: 已完成
	Progress int `json:"progress"` // 任务进度
}

type TaskStatusResponse = APIResponse[TaskStatusData]

// EmptyResponse preserves data:null for operations without response data.
type EmptyResponse = APIResponse[*struct{}]

type FileDetailInfo struct {
	AuditStatus   int    `json:"auditStatus"`
	CTime         int64  `json:"ctime"`
	Depth         int    `json:"depth"`
	DirType       int    `json:"dirType"`
	Ext           string `json:"ext,omitempty"`
	FileID        string `json:"fileId"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize,omitempty"`
	FileType      int    `json:"fileType,omitempty"`
	FullParentIDs string `json:"fullParentIds,omitempty"`
	GCID          string `json:"gcid,omitempty"`
	MD5           string `json:"md5,omitempty"`
	MimeType      string `json:"mineType,omitempty"`
	ParentID      string `json:"parentId,omitempty"`
	ResType       int    `json:"resType"`
	Thumbnail     string `json:"thumbnail,omitempty"`
	UTime         int64  `json:"utime"`
}

type PictureInfo struct {
	PreviewURL string `json:"previewUrl,omitempty"`
}

type FileDetailData struct {
	FileInfo FileDetailInfo `json:"fileInfo"`
	Location string         `json:"location"`
	PicInfo  PictureInfo    `json:"picInfo"`
}

type FileDetailResponse = APIResponse[FileDetailData]

// UploadInfoResponse has nil Data while the upload is still pending.
type UploadInfoResponse = APIResponse[*FileDetailInfo]

type DownloadURLData struct {
	RequestID        string `json:"requestId"`
	SignedURL        string `json:"signedURL"`
	SpeedupSignature string `json:"speedupSignature"`
	URLDuration      int    `json:"urlDuration"`
}

type DownloadURLResponse = APIResponse[DownloadURLData]

type ShareCreateData struct {
	Code       string `json:"code"`
	CreateTime string `json:"createTime"`
	ShareID    string `json:"shareId"`
	ShareURL   string `json:"shareUrl"`
}

type ShareCreateResponse = APIResponse[ShareCreateData]

type ShareFilesSizeData struct {
	TotalSize   int64            `json:"totalSize"`
	FileSizeMap map[string]int64 `json:"fileSizeMap"`
}

type ShareFilesSizeResponse = APIResponse[ShareFilesSizeData]

type ShareDownloadURLData struct {
	DownloadURL string `json:"downloadUrl"`
}

type ShareDownloadURLResponse = APIResponse[ShareDownloadURLData]

type CloudTask struct {
	CreateTime    int64  `json:"createTime"`
	ErrCode       int    `json:"errCode,omitempty"`
	ErrMsg        string `json:"errMsg,omitempty"`
	FileName      string `json:"fileName"`
	IsDir         bool   `json:"isDir"`
	ParentDirType int    `json:"parentDirType"`
	Res           string `json:"res"`
	ResType       int    `json:"resType"`
	Status        int    `json:"status"`
	TaskID        string `json:"taskId"`
	TotalSize     int64  `json:"totalSize"`
}

type CloudStatusCount struct {
	Status int `json:"status"`
	Count  int `json:"count,omitempty"`
}

type CloudTaskListData struct {
	Cursor       string             `json:"cursor"`
	List         []CloudTask        `json:"list"`
	StatusCounts []CloudStatusCount `json:"statusCounts"`
	Total        int                `json:"total"`
}

type CloudTaskListResponse = APIResponse[CloudTaskListData]

type CloudSubfile struct {
	FileName  string         `json:"fileName"`
	FileIndex *int           `json:"fileIndex,omitempty"`
	FileSize  int64          `json:"fileSize,omitempty"`
	IsDir     bool           `json:"isDir,omitempty"`
	Subfiles  []CloudSubfile `json:"subfiles,omitempty"`
}

type BTResourceInfo struct {
	InfoHash       string         `json:"infoHash"`
	FileName       string         `json:"fileName"`
	FileSize       int64          `json:"fileSize"`
	SubfilesNum    int            `json:"subfilesNum"`
	Subfiles       []CloudSubfile `json:"subfiles"`
	CreateTime     int64          `json:"createTime,omitempty"`
	ExcludeIndices []int          `json:"excludeIndices,omitempty"`
}

type CloudResolveURLData struct {
	ResType   int            `json:"resType"`
	BTResInfo BTResourceInfo `json:"btResInfo"`
	URL       string         `json:"url"`
}

type CloudResolveURLResponse = APIResponse[CloudResolveURLData]

type CloudResolveTorrentData struct {
	ResType   int            `json:"resType"`
	BTResInfo BTResourceInfo `json:"btResInfo"`
}

type CloudResolveTorrentResponse = APIResponse[CloudResolveTorrentData]

type CloudCreateTaskData struct {
	TaskID string `json:"taskId"`
	URL    string `json:"url"`
}

type CloudCreateTaskResponse = APIResponse[CloudCreateTaskData]

type ShareListData struct {
	List  []ShareItem `json:"list,omitempty"`
	Total int         `json:"total,omitempty"`
}

type ShareItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Code             string `json:"code,omitempty"`
	ShareURL         string `json:"shareUrl,omitempty"`
	AccessToken      string `json:"accessToken,omitempty"`
	ValidateDuration int    `json:"validateDuration,omitempty"`
	CreateTime       int64  `json:"createTime,omitempty"`
}

type ShareListResponse = APIResponse[ShareListData]

type UploadCredentials struct {
	AccessKeyID     string     `json:"accessKeyID"`
	SecretAccessKey string     `json:"secretAccessKey"`
	SessionToken    string     `json:"sessionToken"`
	Expiration      *time.Time `json:"expiration,omitempty"`
}

type UploadTokenData struct {
	GCID         string            `json:"gcid,omitempty"`
	Provider     int               `json:"provider,omitempty"`
	TaskID       string            `json:"taskId"`
	Creds        UploadCredentials `json:"creds"`
	Endpoint     string            `json:"endPoint,omitempty"`
	FullEndpoint string            `json:"fullEndPoint"`
	BucketName   string            `json:"bucketName"`
	ObjectPath   string            `json:"objectPath"`
	CallbackVar  string            `json:"callbackVar,omitempty"`
	Region       string            `json:"region,omitempty"`
}

type UploadTokenResponse = APIResponse[UploadTokenData]

type FlashUploadData struct {
	CanFlashUpload bool `json:"canFlashUpload"`
}

type FlashUploadResponse = APIResponse[FlashUploadData]

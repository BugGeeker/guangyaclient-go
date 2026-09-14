package guangyaclient

const clientID = "aMe-8VSlkrbQXpUR"

var FileType = map[string]int{
	"图片":   1,
	"视频":   2,
	"音频":   3,
	"文档":   4,
	"压缩包":  5,
	"安装包":  8,
	"BT种子": 9,
}

var FileTypeName = func() map[int]string {
	result := make(map[int]string, len(FileType))
	for name, value := range FileType {
		result[value] = name
	}
	return result
}()

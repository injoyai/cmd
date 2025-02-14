package resource

var Exclusive = MResource{
	"in":        {Local: "in", Remote: "in_linux_amd64", RemoteArm: "in_linux_arm", RemoteArm64: "in_linux_arm64"},
	"upgrade":   {Local: "in_upgrade", Remote: "upgrade_linux_amd64", RemoteArm: "upgrade_linux_arm", RemoteArm64: "upgrade_linux_arm64", Key: []string{"in_upgrade"}},
	"forward":   {Local: "forward", Remote: "forward_linux_amd64", RemoteArm: "forward_linux_arm", RemoteArm64: "forward_linux_arm64"},
	"edge":      {Local: "edge", Remote: "edge_linux_amd64", RemoteArm: "edge_linux_arm", RemoteArm64: "edge_linux_arm64"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_linux_amd64", RemoteArm: "edge_mini_linux_arm", RemoteArm64: "edge_mini_linux_arm64"},
	"notice":    {Local: "notice", Remote: "notice_linux_amd64", RemoteArm: "notice_linux_arm", RemoteArm64: "notice_linux_arm64"},
}

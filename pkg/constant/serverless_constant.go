package constant

var (
	BaseImage = MasterIp + ":5000" + "/baseserver"
)

type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var Auth = AuthConfig{
	Username: "myuser",
	Password: "123",
}

var AuthCode = "myuser:123"

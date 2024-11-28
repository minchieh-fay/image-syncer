package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"image-syncer/pkg/utils/types"

	"github.com/sirupsen/logrus"

	"image-syncer/pkg/utils"

	"gopkg.in/yaml.v2"
)

// Config information of sync client
type Config struct {
	// the authentication information of each registry
	AuthList map[string]types.Auth `json:"auth" yaml:"auth"`

	// a <source_repo>:<dest_repo> map
	ImageList map[string]interface{} `json:"images" yaml:"images"`

	// only images with selected os can be sync
	osFilterList []string
	// only images with selected architecture can be sync
	archFilterList []string
}

// NewSyncConfig creates a Config struct
func NewSyncConfig(configFile, authFilePath string, imagelist map[string]string,
	osFilterList, archFilterList []string, logger *logrus.Logger) (*Config, error) {

	if len(configFile) == 0 && len(authFilePath) == 0 {
		logger.Warnf("[Warning] No authentication information found because neither " +
			"config.json nor auth.json provided, image-syncer may not work fine.")
	}

	var config Config
	{
		// authFilePath文件的格式如下
		// {
		// 	"auths": {
		// 					"10.35.8.1:7080": {
		// 									"auth": "YWRtaW46YWRtaW4xMjM=" //base64编码  格式为   username:password
		// 					},
		// 					"10.35.9.161:82": {
		// 									"auth": "YWRtaW46YWRtaW4xMjM="
		// 					}
		// 		}
		// }
		// 将authFilePath文件内容解析到config.AuthList
		type AuthInfo struct {
			Auth string `json:"auth"`
		}
		type Auths struct {
			Auths map[string]AuthInfo `json:"auths"`
		}
		authInfo := Auths{}
		// 下面不要用openAndDecode, 直接用json做decode
		file, err := os.OpenFile(authFilePath, os.O_RDONLY, 0666)
		if err != nil {
			return nil, fmt.Errorf("open file %v error: %v", authFilePath, err)
		}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&authInfo); err != nil {
			return nil, fmt.Errorf("unmarshal auth file error: %v", err)
		}
		config.AuthList = make(map[string]types.Auth)
		for registry, auth := range authInfo.Auths {
			// auth需要base64解码 出 username:password
			authStr := auth.Auth
			authBytes, err := base64.StdEncoding.DecodeString(authStr)
			if err != nil {
				return nil, fmt.Errorf("base64 decode auth error: %v", err)
			}
			authStr = string(authBytes)
			// 切割出 username 和 password
			authArr := strings.Split(authStr, ":")
			if len(authArr) != 2 {
				return nil, fmt.Errorf("auth format error: %v", authStr)
			}

			config.AuthList[registry] = types.Auth{
				Username: authArr[0],
				Password: authArr[1],
				Insecure: true,
			}
		}
	}

	{
		if config.ImageList == nil {
			config.ImageList = make(map[string]interface{})
		}
		for key, value := range imagelist {
			config.ImageList[key] = value
		}

		// if err := openAndDecode(imageFilePath, &config.ImageList); err != nil {
		// 	return nil, fmt.Errorf("decode image file %v error: %v", imageFilePath, err)
		// }
	}

	config.osFilterList = osFilterList
	config.archFilterList = archFilterList

	return &config, nil
}

// Open json file and decode into target interface
func openAndDecode(filePath string, target interface{}) error {
	if !strings.HasSuffix(filePath, ".yaml") &&
		!strings.HasSuffix(filePath, ".yml") &&
		!strings.HasSuffix(filePath, ".json") {
		return fmt.Errorf("only one of yaml/yml/json format is supported")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %v not exist: %v", filePath, err)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("open file %v error: %v", filePath, err)
	}

	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(target); err != nil {
			return fmt.Errorf("unmarshal config error: %v", err)
		}
	} else {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(target); err != nil {
			return fmt.Errorf("unmarshal config error: %v", err)
		}
	}

	fmt.Printf("%v-------------------config file %v loaded successfully\n", filePath, target)

	return nil
}

// GetAuth gets the authentication information in Config
func (c *Config) GetAuth(repository string) (types.Auth, bool) {
	auth := types.Auth{}
	prefixLen := 0
	exist := false

	for key, value := range c.AuthList {
		if matched := utils.RepoMathPrefix(repository, key); matched {
			if len(key) > prefixLen {
				auth = value
				exist = true
			}
		}
	}

	return auth, exist
}

func expandEnv(authMap map[string]types.Auth) map[string]types.Auth {
	result := make(map[string]types.Auth)

	for registry, auth := range authMap {
		pwd := os.ExpandEnv(auth.Password)
		name := os.ExpandEnv(auth.Username)
		newAuth := types.Auth{
			Username: name,
			Password: pwd,
			Insecure: true,
		}
		result[registry] = newAuth
	}

	return result
}

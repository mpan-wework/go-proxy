package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func handleRequest(orig *http.Request) string {
	regex := regexp.MustCompile(`^.*?/proxy/([^\/]+)/`)
	u, err := url.Parse(string(regex.ReplaceAll(
		[]byte(orig.URL.Path), []byte("$1://"),
	)))
	if err != nil {
		return ""
	}

	req, _ := http.NewRequest(orig.Method, u.String(), orig.Body)

	for queryKey, queryValues := range orig.URL.Query() {
		for _, queryValue := range queryValues {
			req.URL.Query().Add(queryKey, queryValue)
		}
	}

	for headerKey, headerValues := range orig.Header {
		for _, headerValue := range headerValues {
			req.Header.Add(headerKey, headerValue)
		}
	}
	req.Proto = orig.Proto
	fmt.Println(req.Method, req.URL)
	client := http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	return string(body)
}

func dynamicHandleFunc(c *gin.Context) {
	c.String(http.StatusOK, handleRequest(c.Request))
}

func initConfig() {
	viper.SetConfigFile("config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("config error: %s", err))
	}
}

func main() {
	initConfig()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	v1 := r.Group("/proxy")
	v1.GET("/*dynamic", dynamicHandleFunc)
	v1.POST("/*dynamic", dynamicHandleFunc)
	v1.PUT("/*dynamic", dynamicHandleFunc)
	v1.PATCH("/*dynamic", dynamicHandleFunc)
	v1.DELETE("/*dynamic", dynamicHandleFunc)

	r.Run(fmt.Sprintf(":%v", viper.Get("web.port")))
}

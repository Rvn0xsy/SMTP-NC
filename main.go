package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

//go:embed templates public
var fsEmbed embed.FS
var address = ":8099"
var isHelp bool
type ProcessData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	SMTPFrom string `json:"smtp_from"`
	Port string `json:"port"`
	To []string `json:"to"`
	Body string `json:"body"`
}

func SendSMTP(Data * ProcessData){
	auth := smtp.PlainAuth("", Data.Username, Data.Password, Data.Hostname)
	msg := []byte(Data.Body)
	err := smtp.SendMail(Data.Hostname +":" + Data.Port, auth, Data.SMTPFrom, Data.To, msg)
	if err != nil {
		log.Println(err)
	}else {
		log.Println("Send Success!")
	}
}

func init() {
	flag.StringVar(&address, "listen", ":8099", "listen  address (:8099)")
	flag.BoolVar(&isHelp, "h",false,"help usage")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "SMTP NetCat help %s:\n", os.Args[0])

		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stdout, "-%v\t%v\n", f.Name,f.Usage)
		})
	}
}

func main() {
	flag.Parse()

	if isHelp {
		flag.Usage()
		return
	}

	r := gin.Default()
	must := template.Must(template.New("").ParseFS(fsEmbed, "templates/*.html"))

	fe, _ := fs.Sub(fsEmbed, "public")
	r.StaticFS("/public", http.FS(fe))

	r.SetHTMLTemplate(must)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html",gin.H{})
	})

	r.POST("/send", func(c *gin.Context) {
		var Data ProcessData
		err := c.ShouldBindJSON(&Data)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		go SendSMTP(&Data)
		c.String(200, "OK")
	})

	err := r.Run(address)
	if err != nil {
		log.Fatal(err)
	}

}


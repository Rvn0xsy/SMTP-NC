package main

import (
	"crypto/tls"
	"embed"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"log"
	"net"
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
	EnableSSL bool `json:"enable_ssl"`
}

func SendSMTP(Data * ProcessData){
	auth := smtp.PlainAuth("", Data.Username, Data.Password, Data.Hostname)
	msg := []byte(Data.Body)
	servername := Data.Hostname +":" + Data.Port
	if !Data.EnableSSL {
		err := smtp.SendMail(servername, auth, Data.SMTPFrom, Data.To, msg)
		if err != nil {
			log.Println(err)
		}else {
			log.Println("Send Success!")
		}
	}else{
		host, _, _ := net.SplitHostPort(servername)
		// TLS config
		tlsconfig := &tls.Config {
			InsecureSkipVerify: true,
			ServerName: host,
		}
		conn, err := tls.Dial("tcp", servername, tlsconfig)
		if err != nil {
			log.Println(err)
			return
		}
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			log.Println(err)
			return
		}

		// Auth
		if err = c.Auth(auth); err != nil {
			log.Println(err)
			return
		}

		// To && From
		if err = c.Mail(Data.SMTPFrom); err != nil {
			log.Println(err)
			return
		}

		if err = c.Rcpt(Data.To[0]); err != nil {
			log.Println(err)
			return
		}

		// Data
		w, err := c.Data()
		if err != nil {
			log.Println(err)
			return
		}
		_, err = w.Write(msg)
		if err != nil {
			log.Println(err)
			return
		}
		err = w.Close()
		if err != nil {
			log.Println(err)
			return
		}
		c.Quit()
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


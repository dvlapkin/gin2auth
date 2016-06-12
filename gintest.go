package main
import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"time"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	//reg "gitlab.mobile-info.ru/lap/mailreg"
)
type Person struct{
	Id	int
	Uuid	int
	Login	string
	Status	bool
	Pass	string
	Email	string
	Key	string
	Datereg	time.Time
}
type email_auth struct{
	Login	string
	Host	string
	Pass	string
}

type Au_resp struct{
	Success bool
	Expires int64
	Login string
}
var (	db *sql.DB
	db_au string = "user=lap password=qwerty dbname=test1 host=localhost sslmode=require"
	err error)

///////////////////////////////////

func main() {
	db, err = sql.Open("postgres", db_au)
	if err != nil {
		log.Fatal(err)}
	defer db.Close()

	err=db.Ping()
	if err != nil {
		fmt.Println("no ping to db")
	}else { fmt.Println("db ping ok")}

	fmt.Println("hi!")

	// gin router
  router := gin.Default()
router.LoadHTMLGlob("/home/lap/go/templates/*")
		router.GET("/",index)
	  router.GET("/someGet", getting)
		router.GET("/auth/regconfirm", Fregconfirm)
	  router.POST("/auth/register", Freg)
		router.POST("/auth/login", Flogin)
  // Listen and server on 0.0.0.0:8080
    router.Run(":8080")
}
////////////
func pingdb(){
	err=db.Ping()
	if err != nil{
		fmt.Println("no ping to db")
		}else{
		fmt.Println("db ping ok")}
}

///////////////////////////////////////////////////////////
func index(c *gin.Context){
	var hello=" who a u?"
	fmt.Println("getting")
	a,l:= auth(c)
	if a {hello=l}
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"name": hello,})
}

func getting(c *gin.Context){
	res,l := auth(c)
	if res {fmt.Println(l)}
	c.String(200,"pong")
	fmt.Println("getting")
}

/////////////////////////////////////////////////////////////
func Freg(c *gin.Context){
	fmt.Println("registration")
	var user Person
	user.Uuid = 0
	user.Login = c.PostForm("login")
	user.Status = false
	user.Pass = c.PostForm("password")
	user.Email = c.PostForm("email")
	user.Key=JwToken(user.Login,string(user.Uuid))

	if db != nil{
		pingdb()
		}else{
			fmt.Println("no db")
			c.JSON(200, gin.H{"success":"false"})
			return}

	if (user.Login!="")&&(user.Pass!="")&&(user.Email!=""){
		fmt.Println("reg data ok")
		//prep transaction
		Tx,txerr := db.Begin()
		if txerr != nil {
			fmt.Println("begin err")
			return}

	/*	stmt,err := Tx.Prepare("INSERT INTO users(login) VALUES ($1)")
			if err != nil {
				fmt.Println("prep1 err")
				return}
*/
		stmt,err := Tx.Prepare("INSERT INTO users(uuid,login,status,password,email,key,datereg) VALUES ($1,$2,$3,$4,$5,$6,$7)")
		if err != nil {
			fmt.Println("prep2 err")
			return}
		//insert
		_, err = stmt.Exec(user.Uuid,user.Login,user.Status,user.Pass,user.Email,user.Key,time.Now())
		if err != nil {
			fmt.Println("exec err")
			return}
		//
		err = Tx.Commit()
		if err != nil {
			fmt.Println("commit err")
			return}

		c.JSON(200, gin.H{"success":"true"})
		sendmail(&user)
	}else{
		fmt.Println("reg err 1")
		c.JSON(200, gin.H{"success":"false"})
	}
}

///////////////////////////////////

func JwToken(a,b string) string{
    token := jwt.New(jwt.SigningMethodHS256)
    token.Claims["login"] = a
    token.Claims["uuid"] = b
    tokenString, err := token.SignedString([]byte("sign"))
	if err != nil {
		log.Fatal(err)}
return tokenString
}

///////////////////////////////////

func sendmail(user *Person){
	fmt.Println("sending mail...")

	cfg,err := ioutil.ReadFile("/home/lap/go/src/gitlab.mobile-info.ru/lap/gintest/email.cfg")
    	if err != nil {
		log.Panic(err)
       	 	fmt.Println("read email cfg err")
		return}
	dec := json.NewDecoder(strings.NewReader(string(cfg)))
	var mail email_auth
	err = dec.Decode(&mail);
	if err != nil {
			fmt.Println("decode email cfg err")
			return}
	mm := mail.Login+"@"+mail.Host
	host := "smtp."+mail.Host
	TLS_server := host+":465"
	//smtp_server := host+":25"
	//_,err := smtp.Dial(smtp_server)

	tlsconfig := &tls.Config {
        	InsecureSkipVerify: true,
        	ServerName: TLS_server,
   		}
    	conn, err := tls.Dial("tcp",TLS_server, tlsconfig)
    	if err != nil {
       	 	log.Panic(err)
  	}
	defer conn.Close()
	fmt.Println("dial ok...")

	client, err := smtp.NewClient(conn, host)
    	if err != nil {
        	log.Panic(err)}
/*
   	err = client.StartTLS(tlsconfig)
    	if err != nil {
       	 	log.Panic(err)
  	}
	fmt.Println("TLS ok...")
*/
	auth := smtp.PlainAuth("", mm,mail.Pass, host)
	if err = client.Auth(auth); err != nil {
        	log.Panic(err)
    	}
	fmt.Println("auth ok...")

	 // To && From
    	if err = client.Mail(mm); err != nil {
        	log.Panic(err)}
    	if err = client.Rcpt(mm); err != nil {
        	log.Panic(err)}

    	// Data
    	w, err := client.Data()
    	if err != nil {
        	log.Panic(err)
    	}

	mt := template.Must(template.ParseFiles("/home/lap/go/src/gitlab.mobile-info.ru/lap/gintest/email.html"))
    	if err != nil {
        	log.Panic(err)
    	}

	fmt.Println("parse ok...")
	err = mt.Execute(w,&user)
	if err != nil {
        	log.Panic(err)
    	}
	fmt.Println("exec temp ok...")

    	err = w.Close()
    	if err != nil {
        	log.Panic(err)
    	}
    	client.Quit()
	fmt.Println("sent!")
}

////////////////////////////////////

func Fregconfirm(c *gin.Context){
	fmt.Println("confirm..")
	var user Person
	user.Login = c.Query("login")
	user.Key =c.Query("key")
	user.Uuid,_ = strconv.Atoi(c.Query("uuid"))

pingdb()

	Tx,txerr := db.Begin()
	if txerr != nil {
		fmt.Println("begin err")
		return}

	stmt,err := Tx.Prepare("UPDATE users SET status = TRUE where login=$1 AND uuid=$2 AND key=$3")
	if err != nil {
		log.Fatal(err)}

	_, err = stmt.Exec(user.Login,user.Uuid,user.Key)
	if err != nil {
		fmt.Println("update error")
		log.Fatal(err)}

	err = Tx.Commit()
		if err != nil {
			log.Fatal(err)}
	fmt.Println("login from url ="+user.Login)
	fmt.Println("key from url ="+user.Key)
	fmt.Println("confirmed!")
	c.JSON(200, gin.H{"success":"true"})
}

///////////////////////////

func auth(c *gin.Context) (res bool,login string) {
	//r,err :=http.ReadRequest()
	fmt.Println("*** authorization...")
	r:=c.Request
//	fmt.Println("*req Headers:", c.Request.Header)
/*	if err != nil {
		fmt.Println("readRequest err")
		login=""
		res=false
		return}
*/
	coologin, err := r.Cookie("login")
	if err != nil {
		fmt.Println("au cookie err")
		login=""
		res=false
		return}
	login=coologin.Value

	jwt, err := r.Cookie("jwt")
	if err != nil {
		fmt.Println("au cookie err")
		res=false
		return}
	uuid, err := r.Cookie("uuid")
	if err != nil {
		fmt.Println("au cookie err")
		res=false
		return}

	u,err 	:= strconv.Atoi(uuid.Value)
	jj	:= JwToken(login,string(u))
	if jj != jwt.Value {
		//c.JSON(200, gin.H{"success":"false"})
		fmt.Println("*** bad authorization")
		res=false
		return
		}
res=true
fmt.Println(login)
fmt.Println("*** authorization ok!")
return
}

////////////////////////////////////

func Flogin(c *gin.Context){
	w := c.Writer
//coockSet(w,"Jj","Jjj",32)
	fmt.Println("*** login...")
	post_login:= c.PostForm("login")
	post_pass:= c.PostForm("password")
	if (post_login!="")&&(post_pass!=""){
		fmt.Println(post_login,post_pass)
		row:= db.QueryRow("SELECT key,uuid FROM users WHERE login=$1 AND password=$2",post_login,post_pass)
		var k string
		var u int

		err := row.Scan(&k,&u)
		if err != nil {
			coockSet(w,"","",0)
			fmt.Println("*** bad login ")
			c.JSON(200, gin.H{"success":"false"})
			return
		}else{
			fmt.Println("*** login ok")
			coockSet(w,k,post_login,u)
			resp2 := Au_resp{Expires:0,Success:true,Login:post_login}
			fmt.Println(resp2)
			c.JSON(200, gin.H{"success":"true"})
			return
		}

	}else {
		coockSet(w,"","",0)
		c.JSON(200, gin.H{"success":"false"})
		return
	}
	//c.JSON(200, gin.H{"success":"true"})
}

//////////////////////////////////////

func coockSet(w http.ResponseWriter, j,l string, u int){
	fmt.Println("Cookie set:");
	cook := http.Cookie{Name: "jwt", Value: j,Path:"/"}
	fmt.Println(cook.String())
	http.SetCookie(w,&cook)
	cook = http.Cookie{Name: "login", Value: l,Path:"/"}
	fmt.Println(cook.String())
	http.SetCookie(w,&cook)
	cook = http.Cookie{Name: "uuid", Value: strconv.Itoa(u),Path:"/"}
	fmt.Println(cook.String())
	http.SetCookie(w,&cook)
}

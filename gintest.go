package main
import ("fmt"
	"database/sql"
	"io/ioutil"
	"log"
	"time"
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

var (	db *sql.DB
	db_au string = "user=lap password=qwerty dbname=test1 host=localhost sslmode=require"
	err error)

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
	router.GET("/",index)
   	router.GET("/someGet", getting)
    	router.POST("/auth/register", Freg)
	
    // Listen and server on 0.0.0.0:8080
    	router.Run(":8080")
}

///////////////////////////////////////////////////////////
func index(c *gin.Context){

	fmt.Println("getting")
	indx,err := ioutil.ReadFile("../src/gitlab.mobile-info.ru/lap/gintest/index.html")
	if err != nil{
		fmt.Println("error N 1")}
	c.JSON(200, gin.H{"status":"ok","name":string(indx)})
}

func getting(c *gin.Context){
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

	if db != nil {
		err=db.Ping()
		if err != nil {
			fmt.Println("no ping to db")
		}else { fmt.Println("db ping ok")}
	}else {fmt.Println("no ping db")
		c.JSON(200, gin.H{"success":"false"})
		return}

	if (user.Login!="")&&(user.Pass!="")&&(user.Email!=""){
		fmt.Println("reg data ok")
		//prep transaction
		Tx,txerr := db.Begin()
		if txerr != nil {
			fmt.Println("begin err")
			return}	
		stmt,err := Tx.Prepare("INSERT INTO users(uuid,login,status,password,email,key,datereg) VALUES ($1,$2,$3,$4,$5,$6,$7)")
		if err != nil {
			fmt.Println("prep err")
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
	}else{
		fmt.Println("reg err 1")
		c.JSON(200, gin.H{"success":"false"})
	}
}

func JwToken(a,b string) string{
    token := jwt.New(jwt.SigningMethodHS256)
    token.Claims["login"] = a
    token.Claims["uuid"] = b
    tokenString, err := token.SignedString([]byte("sign"))
	if err != nil {
		log.Fatal(err)}
return tokenString
}

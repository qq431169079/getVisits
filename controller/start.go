package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	userName  string = "XXX"
	password  string = "XXX"
	ipAddrees string = "123.206.224.239"
	port      int    = 3306
	dbName    string = "visits"
	charset   string = "utf8"
)

var secrets = gin.H{
	"admin": gin.H{"name": "cq", "password": "admin12345"},
}

func connectMysql() *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", userName, password, ipAddrees, port, dbName, charset)
	Db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("mysql connect failed, detail is [%v]", err.Error())
	}
	return Db
}
func addRecord(Db *sqlx.DB, ip string, date string, address string, referer string) {
	result, err := Db.Exec("insert into info (ip,address,visitdate,referer) values(?,?,?,?)", ip, address, date, referer)
	if err != nil {
		fmt.Printf("data insert faied, error:[%v]", err.Error())
		return
	}
	id, _ := result.LastInsertId()
	fmt.Printf("insert success, last id:[%d]\n", id)
}

type userInfo struct {
	Id        int            `db:"id"`
	Ip        sql.NullString `db:"ip"`
	Address   sql.NullString `db:"address"`
	Visitdate string         `db:"visitdate"`
}

// 查询结果集
func queryData(Db *sqlx.DB, date string) (result []interface{}) {
	rows, err := Db.Query("select * from info where visitdate = ?", date)
	if err != nil {
		fmt.Printf("query faied, error:[%v]", err.Error())
		return
	}
	var arrays = make([]interface{}, 0)

	for rows.Next() {
		//定义变量接收查询数据
		var id int
		var ip string
		var address sql.NullString
		var visitdate string
		var referer sql.NullString

		err := rows.Scan(&id, &ip, &address, &visitdate, &referer)
		if err != nil {
			fmt.Println("get data failed, error:[%v]", err.Error())
		}
		log.Println(id, ip, address, visitdate)

		var infoMap map[string]string /*创建集合 */
		infoMap = make(map[string]string)
		infoMap["ip"] = ip
		infoMap["address"] = address.String
		infoMap["visitdate"] = visitdate
		infoMap["referer"] = referer.String
		arrays = append(arrays, infoMap)
		log.Println(arrays)
	}
	//关闭结果集（释放连接）
	rows.Close()
	return arrays
}

// 查询单个数据
func getData(Db *sqlx.DB, date string) (info userInfo) {
	//初始化定义结构体，用来存放查询数据
	var userData *userInfo = new(userInfo)
	err := Db.Get(userData, "select * from info where visitdate = ?", date)
	if err != nil {
		fmt.Printf("query faied, error:[%v]", err.Error())
		return
	}
	//打印结构体内容
	fmt.Println(userData)
	return *userData
}

func main() {
	r := gin.Default()

	authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"admin": "admin12345",
	}))

	r.StaticFS("static", http.Dir("static"))

	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	r.GET("/visits", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// ip := c.Request.RemoteAddr 如果是这种方式获取不到真实ip
		// 获取真实ip
		ip := c.ClientIP()
		url := c.Request.URL.String()
		referer := c.Request.Referer()
		fmt.Println("请求ip：", ip)
		fmt.Println("请求url：", url)
		fmt.Println("请求资源：", referer)
		// cut := strings.Index(ip, ":")
		// ipNoPort := ip[0:cut]
		result, _ := http.Get("http://apis.haoservice.com/lifeservice/QueryIpAddr/query?ip=" + ip + "&key=XXXXXX");
		body, _ := ioutil.ReadAll(result.Body)
		fmt.Println("接口返回信息：", string(body))
		// 接口返回信息转map
		var resultMap map[string]interface{} /*创建集合 */
		resultMap = make(map[string]interface{})
		err := json.Unmarshal([]byte(body), &resultMap)
		if err != nil {
			fmt.Println(err)
		}
		infoMap := resultMap["result"].(map[string]interface{})

		fmt.Println("国家：", infoMap["country"])
		fmt.Println("省份：", infoMap["province"])
		fmt.Println("城市：", infoMap["city"])
		fmt.Println("服务商：", infoMap["isp"])
		var country string
		var province string
		var city string
		var isp string
		var address string
		if infoMap["country"] != nil {
			country = infoMap["country"].(string)
		}
		if infoMap["province"] != nil {
			province = infoMap["province"].(string)
		}
		if infoMap["city"] != nil {
			city = infoMap["city"].(string)
		}
		if infoMap["isp"] != nil {
			isp = infoMap["isp"].(string)
		}
		address = country + province + city + isp

		log.Println(address)

		timeUnix := time.Now().Unix()
		date := time.Unix(timeUnix, 0).Format("2006-01-02")
		log.Println(date)

		var Db *sqlx.DB = connectMysql()
		addRecord(Db, ip, date, address, referer)
		defer Db.Close()
		c.String(200, "ok")
	})

	authorized.GET("/getVisits", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		user := c.MustGet(gin.AuthUserKey).(string)
		if _, ok := secrets[user]; ok {
			//date, _ := strconv.Atoi(c.Query("date"))
			date := c.Query("date")
			var Db *sqlx.DB = connectMysql()
			data := queryData(Db, date)
			log.Println(data)
			defer Db.Close()
			c.JSON(http.StatusOK, gin.H{"date": data})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": "没有权限 :("})
		}
	})
	r.Run(":8887")
}

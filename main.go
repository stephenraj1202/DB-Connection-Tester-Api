package main

import (
	"context"
	"corent/log"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var db *sql.DB
var err error
var ip, username, password, port string

func main() {
	router := mux.NewRouter()
	fmt.Println("Listeneing on Port:" + "9000")
	router.HandleFunc("/", IndexPage).Methods(http.MethodGet)
	router.HandleFunc("/mysql", mysqlDb).Methods(http.MethodGet)
	router.HandleFunc("/mysqlforminput", mysqlforminput).Methods(http.MethodGet)
	router.HandleFunc("/mysqldatabases/{dbname}", mysqldatabases).Methods(http.MethodGet)
	router.HandleFunc("/mysqldatabases/mysqltables/{dbname}/{table}", mysqltables).Methods(http.MethodGet)
	router.HandleFunc("/mysqldatabases/mysqlqueryexecute/{dbname}", mysqlqueryexecute).Methods(http.MethodGet)
	router.HandleFunc("/mysqlqueryGetting", mysqlqueryGetting).Methods(http.MethodGet)
	router.HandleFunc("/mysqlqueryexecution", mysqlqueryexecution).Methods(http.MethodGet)
	router.HandleFunc("/mysqldatabases/mysqldump/{dbname}", mysqldump).Methods(http.MethodGet)
	router.HandleFunc("/mysqlprovisioning", mysqlprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/mysqldeprovisioning", mysqldeprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/mssql", mssql).Methods(http.MethodGet)
	router.HandleFunc("/mssqlforminput", mssqlforminput).Methods(http.MethodGet)
	router.HandleFunc("/mssqldatabases/{dbname}", mssqldatabases).Methods(http.MethodGet)
	router.HandleFunc("/mssqldatabases/mssqldump/{dbname}", mssqldump).Methods(http.MethodGet)
	router.HandleFunc("/mssqlprovisioning", mssqlprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/mssqldeprovisioning", mssqldeprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/mssqldatabases/mssqltables/{dbname}/{table}", mssqltables).Methods(http.MethodGet)
	router.HandleFunc("/postgresql", postgresql).Methods(http.MethodGet)
	router.HandleFunc("/postgreforminput", postgresqlinput).Methods(http.MethodGet)
	router.HandleFunc("/postgredatabases/{dbname}", postgredatabases).Methods(http.MethodGet)
	router.HandleFunc("/postgredatabases/postgretables/{dbname}/{table}", postgretables).Methods(http.MethodGet)
	router.HandleFunc("/postgredatabases/postgredump/{dbname}", postgredump).Methods(http.MethodGet)
	router.HandleFunc("/postgreprovisioning", postgreprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/postgredeprovisioning", postgredeprovisioning).Methods(http.MethodGet)
	router.HandleFunc("/sqlite3", sqlite3DB).Methods(http.MethodGet)
	router.HandleFunc("/sqliteforminput", sqliteforminput).Methods(http.MethodGet)
	router.HandleFunc("/sqlitetables/{table}", sqlitetables).Methods(http.MethodGet)
	router.HandleFunc("/mongodb", mongodbIndex).Methods(http.MethodGet)
	router.HandleFunc("/mongodforminput", mongodforminput).Methods(http.MethodGet)
	router.HandleFunc("/mongodatabases/{dbname}", mongodatabases).Methods(http.MethodGet)
	router.HandleFunc("/mongodatabases/mongocollection/{dbname}/{collection}", mongocollection).Methods(http.MethodGet)
	_ = http.ListenAndServe(":"+"9000", router)
}

func mssqldeprovisioning(writer http.ResponseWriter, request *http.Request) {
	targetdbname := request.URL.Query().Get("targetdbname")
	targetdbname = strings.TrimSpace(targetdbname)
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;",
		ip, username, password, port)
	db, err = sql.Open("mssql", connString)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	var mskill_ver2k = `USE master;
DECLARE @kill varchar(8000); SET @kill = '';  
SELECT @kill = @kill + 'kill ' + CONVERT(varchar(5), spid) + ';'  
FROM master..sysprocesses ` +
		"\nWHERE dbid  = db_id(" + "'" + targetdbname + "'" + ")\n" + "EXEC(@kill);"
	var killmssqldb = `USE [master];
	DECLARE @kill varchar(8000) = '';
	SELECT @kill = @kill + 'kill ' + CONVERT(varchar(5), session_id) + ';'
	FROM sys.dm_exec_sessions ` +
		"\nWHERE database_id  = db_id(" + "'" + targetdbname + "'" + ")\n" + "EXEC(@kill);"
	rows, err := db.Query("SELECT @@version")
	if err != nil {
		log.Error("error in getting version in mssql", err)
	}
	defer rows.Close()
	var mssqlversion string
	for rows.Next() {
		if err := rows.Scan(&mssqlversion); err != nil {
			log.Error("error in getting version rows in mssql", err)
		}
	}
	if strings.Contains(mssqlversion, "2012") {
		err = executionQuery(killmssqldb)
		if err != nil {
			log.Error("error in execution query", err)
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))
		}
	} else {

		err = executionQuery(mskill_ver2k)
		if err != nil {
			log.Error("error in execution query", err)
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))
		}
	}

	err = executionQuery("DROP DATABASE " + targetdbname)
	if err != nil {
		log.Error("error in execution query", err)
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))

	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
			`<br><h3><a href="/mysql">Back To MySQL Home....</a></h3></body></html>`))
	}
}

func mssqlprovisioning(writer http.ResponseWriter, request *http.Request) {
	targetdbname := request.URL.Query().Get("targetdbname")
	targetdbname = strings.TrimSpace(targetdbname)
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		ip, username, password, port, dbname)

	db, err = sql.Open("mssql", connString)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	tableName, status := getTableNames(db, "SELECT name FROM sys.tables;")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		err = mssqlsqlconvertor(dbname, targetdbname, tableName, db)
		if err != nil {
			log.Error("error in execution query", err)
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))

		} else {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
				`<br><h3><a href="/mysql">Back To MySQL Home....</a></h3></body></html>`))
		}
	}
}

func mssqlsqlconvertor(srcdbname string, targetdbname string, tableNameArr [] string, db *sql.DB) error {
	err = executionQuery("CREATE DATABASE " + targetdbname)
	if err != nil {
		return err
	}
	//select * into test1.dbo.users from jtrac.dbo.users;
	log.Info("srcdbname", srcdbname)
	log.Info("targetdbname", targetdbname)
	log.Info("tableNameArr", tableNameArr)
	for _, value := range tableNameArr {
		_, err := db.Query("select * into " + targetdbname + ".dbo." + value + " " + "from " + dbname + ".dbo." + value + ";")
		if err != nil {
			return err
		}
	}

	return nil
}

func mssqldump(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html>
	<body>
	<h2>Enter Provisionig/Dump-DbName:</h2>
	<form action="/mssqlprovisioning" method="get">
	<label for="targetdbname">Target-DbName:</label><br>
	<input type="text" id="targetdbname" name="targetdbname" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
<h3>-----------------------------------------------------------------------------------</h3><br><br>
<h2>Enter DeProvisionig/Dump-DbName:</h2>
<form action="/mssqldeprovisioning" method="get">
	<label for="targetdbname">DbName:</label><br>
	<input type="text" id="targetdbname" name="targetdbname" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
	</body>
	</html>`))

}

func mysqldeprovisioning(writer http.ResponseWriter, request *http.Request) {
	targetdbname := request.URL.Query().Get("targetdbname")
	targetdbname = strings.TrimSpace(targetdbname)
	dataSourceName := username + ":" + password + "@" + "tcp" + "(" + ip + ":" + port + ")/" + dbname
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	err = executionQuery("DROP DATABASE " + targetdbname)
	if err != nil {
		log.Error("error in execution query", err)
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))

	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
			`<br><h3><a href="/mysql">Back To MySQL Home....</a></h3></body></html>`))
	}
}

func mysqlprovisioning(writer http.ResponseWriter, request *http.Request) {
	targetdbname := request.URL.Query().Get("targetdbname")
	targetdbname = strings.TrimSpace(targetdbname)
	dataSourceName := username + ":" + password + "@" + "tcp" + "(" + ip + ":" + port + ")/" + dbname
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	tableName, status := getTableNames(db, "SELECT table_name FROM information_schema.tables where table_schema="+"'"+dbname+"'")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		err = mysqlsqlconvertor(dbname, targetdbname, tableName, db)
		if err != nil {
			log.Error("error in execution query", err)
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))

		} else {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
				`<br><h3><a href="/mysql">Back To MySQL Home....</a></h3></body></html>`))
		}
	}

}
func mysqlsqlconvertor(srcdbname string, targetdbname string, tableNameArr [] string, db *sql.DB) error {
	err = executionQuery("CREATE DATABASE " + targetdbname)
	if err != nil {
		return err
	}
	//create table targetdb.table select * from jtrac.table
	//create table targetdb.users like jtrac.users;
	//insert into targetdb.users select * from jtrac.users;
	for _, value := range tableNameArr {
		err = executionQuery("create table " + targetdbname + "." + value + " " + "like " + dbname + "." + value)
		if err != nil {
			return err
		}
	}
	for _, value := range tableNameArr {
		err = executionQuery("insert into " + targetdbname + "." + value + " " + "select * from " + dbname + "." + value)
		if err != nil {
			return err
		}
	}
	return nil
}

//<label for="sourcedbname">Source-DbName:</label><br>
//<input type="text" id="sourcedbname" name="sourcedbname" value=""><br><br>

func mysqldump(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html>
	<body>
	<h2>Enter Provisionig/Dump-DbName:</h2>
	<form action="/mysqlprovisioning" method="get">
	<label for="targetdbname">Target-DbName:</label><br>
	<input type="text" id="targetdbname" name="targetdbname" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
<h3>-----------------------------------------------------------------------------------</h3><br><br>
<h2>Enter DeProvisionig/Dump-DbName:</h2>
<form action="/mysqldeprovisioning" method="get">
	<label for="targetdbname">DbName:</label><br>
	<input type="text" id="targetdbname" name="targetdbname" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
	</body>
	</html>`))
}

func postgredeprovisioning(writer http.ResponseWriter, request *http.Request) {
	deprovisioning := request.URL.Query().Get("deprovisioning")
	deprovisioning = strings.TrimSpace(deprovisioning)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ip, port, username, password, "postgres")
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	statement := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname =` + "'" + deprovisioning + "'" + ")"

	row := db.QueryRow(statement)
	var exists bool
	err = row.Scan(&exists)

	if exists {
		disconnectQuery := "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname= '" + "postgres" + "'"
		_, err = db.Query(disconnectQuery)
		if err != nil {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Failed..." + "</h2><br>" + err.Error() +
				`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		}
		_, err = db.Exec("DROP DATABASE " + deprovisioning + ";")
		if err != nil {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Failed..." + "</h2><br>" + err.Error() +
				`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		} else {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
				`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		}
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "DbName Not Exists..." + "</h2><br>" +
			`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
	}
	defer db.Close()
}

func postgreprovisioning(writer http.ResponseWriter, request *http.Request) {
	targetdbname := request.URL.Query().Get("targetdbname")
	targetdbname = strings.TrimSpace(targetdbname)
	sourcedbname := request.URL.Query().Get("sourcedbname")
	sourcedbname = strings.TrimSpace(sourcedbname)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ip, port, username, password, "postgres")
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	statement := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname =` + "'" + targetdbname + "'" + ")"

	row := db.QueryRow(statement)
	var exists bool
	err = row.Scan(&exists)

	if !exists {
		//disconnectQuery := "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname= '" + "postgres" + "'"
		//_, err = db.Query(disconnectQuery)
		//if err != nil {
		//	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Failed..." + "</h2><br>" + err.Error() +
		//		`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		//}
		_, err = db.Exec("CREATE DATABASE " + targetdbname + " WITH TEMPLATE " + sourcedbname)
		if err != nil {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Failed..." + "</h2><br>" + err.Error() +
				`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		} else {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
				`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
		}

	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "DbName Already Exists..." + "</h2><br>" +
			`<br><h3><a href="/postgresql">Back To POSTGRESQL Home....</a></h3></body></html>`))
	}
	defer db.Close()
}

func postgredump(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html>
	<body>
	<h2>Enter Provisionig/Dump-DbName:</h2>
	<form action="/postgreprovisioning" method="get">
	<label for="targetdbname">Target-DbName:</label><br>
	<input type="text" id="targetdbname" name="targetdbname" value=""><br><br>
<label for="sourcedbname">Source-DbName:</label><br>
	<input type="text" id="sourcedbname" name="sourcedbname" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
<h3>-----------------------------------------------------------------------------------</h3><br><br>
<h2>Enter DeProvisionig/Dump-DbName:</h2>
<form action="/postgredeprovisioning" method="get">
	<label for="getquery">DbName:</label><br>
	<input type="text" id="deprovisioning" name="deprovisioning" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
	</body>
	</html>`))
}

func mongocollection(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname := param["dbname"]
	collection := param["collection"]
	connString := "mongodb://" + username + ":" + password + "@" + ip + ":" + port + "/" + "admin"
	//connString := "mongodb://devcom_198:devcom_198@192.168.1.245:27017/devcom_198?maxPoolSize=100;waitQueueMultiple=50;safe=true;retryWrites=true"
	clientOptions := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := client.Database(dbname).Collection(collection)
	cursor, err := col.Find(context.TODO(), bson.D{})
	if err != nil {
		_, _ = writer.Write([]byte("Finding all documents ERROR:" + err.Error()))
		log.Error("error in find collection", err)
		defer cursor.Close(ctx)
	} else {
		c := 0
		var result bson.D
		for cursor.Next(ctx) {
			err := cursor.Decode(&result)
			if err != nil {
				_, _ = writer.Write([]byte("cursor.Next() error:" + err.Error()))
			} else {
				for _, v := range result {
					c = c + 1
					jsonBytes, err := json.MarshalIndent(v, "", "    ")
					if err != nil {
						log.Error("error in marshalling", err)
					} else {
						_, _ = writer.Write(jsonBytes)

					}
					//_, _ = writer.Write([]byte("\n"))
				}

			}
		}
		_, _ = writer.Write([]byte("\n" + "Count of Documents: " + strconv.Itoa(c)))
	}

}

func mongodatabases(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	connString := "mongodb://" + username + ":" + password + "@" + ip + ":" + port + "/" + "admin"
	//connString := "mongodb://devcom_198:devcom_198@192.168.1.245:27017/devcom_198?maxPoolSize=100;waitQueueMultiple=50;safe=true;retryWrites=true"
	clientOptions := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	filter := bson.D{{}}
	d := client.Database(dbname)

	arcol, err := d.ListCollectionNames(context.TODO(), filter)
	if err != nil {
		_, _ = writer.Write([]byte("Error getting Collection_name: \n" + err.Error()))
		return
	} else {
		for key, value := range arcol {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=mongocollection/" + dbname + "/" + value + ">" + strconv.Itoa(key+1) + "." + value + `</a></h3></body></html>`))
		}
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Total Number Of Tables:" + strconv.Itoa(len(arcol)) + "</h2>" + `</body></html>`))
	}

}

func mongodforminput(writer http.ResponseWriter, request *http.Request) {
	ip = request.URL.Query().Get("ip")
	ip = strings.TrimSpace(ip)
	username = request.URL.Query().Get("uname")
	username = strings.TrimSpace(username)
	password = request.URL.Query().Get("password")
	password = strings.TrimSpace(password)
	port = request.URL.Query().Get("port")
	port = strings.TrimSpace(port)
	dbname = request.URL.Query().Get("mdbname")
	dbname = strings.TrimSpace(dbname)
	connString := "mongodb://" + username + ":" + password + "@" + ip + ":" + port + "/" + dbname
	//connString := "mongodb://devcom_198:devcom_198@192.168.1.245:27017/devcom_198?maxPoolSize=100;waitQueueMultiple=50;safe=true;retryWrites=true"
	clientOptions := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Error("could not communicate eith server ", err)
		_, _ = writer.Write([]byte("could not communicate eith server " + err.Error()))
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Connected SuccessFully..." + "</h2>" + `</body></html>`))
	}
	filter := bson.D{{}}
	dbs, err := client.ListDatabaseNames(context.TODO(), filter)
	if err != nil {
		_, _ = writer.Write([]byte("Could not getting Dbnames " + err.Error()))
		return
	}
	result1 := mongodatabaseFraming(dbs)
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html lang="en"><head>
<style>
a:link, a:visited {
  background-color: #FFFF00;
  color: black;
  padding: 8px 4px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
}

a:hover, a:active {
  background-color: orange;
}
</style>
</head>
	<body><h4>` + result1 +
		`</h4></body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><br>` + "<h3>" + "Total Number Of DataBases: " + strconv.Itoa(len(dbs)) + "</h3>" + `</body></html>`))

}

func mongodbIndex(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<center>
<h2>Enter Valid MONGODB Input Data:</h2>
<form action="mongodforminput" method="get">
 <label for="ip">IP/HOST:</label><br>
  <input type="text" id="ip" name="ip" value=""><br><br>
  <label for="uname">User Name:</label><br>
  <input type="text" id="uname" name="uname" value=""><br>
  <label for="password">PassWord:</label><br>
  <input type="text" id="password" name="password" value=""><br><br>
  <label for="port">Port:</label><br>
  <input type="text" id="port" name="port" value=""><br><br>
<label for="mdbname">MongoDB-Name:(Default DbName=admin)</label><br>
  <input type="text" id="mdbname" name="mdbname" value=""><br><br>
  <input type="submit" value="Submit">
</form> </center>
</body>
</html>`))
}
func mongodatabaseFraming(id []string) string {
	t := template.Must(template.New("tmpl").Parse(mongodbtemp))
	tr, _ := os.Create("databaseFraming.txt")
	err := t.Execute(tr, id)
	if err != nil {
		panic(err)
	}
	trc, _ := ioutil.ReadFile("databaseFraming.txt")
	defer tr.Close()
	return string(trc)
}

const mongodbtemp = `
 <ul>
{{range $val := .}}
     <a href="mongodatabases/{{$val}}">{{$val}}</a>	 		
{{end}}
</ul>
`

var dbpathname string

func sqlitetables(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	table := param["table"]
	db, err = sql.Open("sqlite3", dbpathname)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	jsonString, err, count := getJSON("select * from " + table + ";")
	if err != nil {
		_, _ = writer.Write([]byte("error in getting table Data " + err.Error()))
	} else {
		_, _ = writer.Write([]byte( jsonString + "\n\n"))
		_, _ = writer.Write([]byte( "Row Count of Tables: " + strconv.Itoa(count) ))
	}
	defer db.Close()
}

func sqliteforminput(writer http.ResponseWriter, request *http.Request) {
	dbpathname = request.URL.Query().Get("dbname")
	dbpathname = strings.TrimSpace(dbpathname)
	dbname := filepath.Base(dbpathname)
	fmt.Println(dbpathname)
	db, err := sql.Open("sqlite3", dbpathname)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	err = db.Ping()
	if err != nil {
		log.Error("could not communicate eith server ", err)
		_, _ = writer.Write([]byte("could not communicate eith server " + err.Error()))
		return
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "SuccessFully Connected to " + dbname + "</h2>" + `</body></html>`))
	}
	tableName, status := getTableNames(db, "SELECT name FROM sqlite_master WHERE type ='table' AND  name NOT LIKE 'sqlite_%';")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		for key, value := range tableName {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=sqlitetables/" + value + ">" + strconv.Itoa(key+1) + "." + value + `</a></h3></body></html>`))
		}
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Total Number Of Tables:" + strconv.Itoa(len(tableName)) + "</h2>" + `</body></html>`))
	}
	defer db.Close()
}

func sqlite3DB(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<center>
<h2>Enter Valid SQLite3 Input Data:</h2>
<form action="/sqliteforminput" method="get">
   <label for="dbname">DB-Path-Name:</label><br>
  <input type="text" id="dbname" name="dbname" value=""><br><br>
  <input type="submit" value="Submit">
</form> </center>
</body>
</html>`))
}

func mysqlqueryexecution(writer http.ResponseWriter, request *http.Request) {
	exequery := request.URL.Query().Get("exequery")
	exequery = strings.TrimSpace(exequery)
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "DBName:" + dbname + "</h4>" + `</body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "Enterd Executing Query:" + exequery + "</h4>" + `</body></html>`))
	err = executionQuery(exequery)
	if err != nil {
		log.Error("error in execution query", err)
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))

	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Query Executed Successfully..." + "</h2>" +
			`<br><h3><a href="/mysql">Back To MySQL Home....</a></h3></body></html>`))
	}

}

func mysqlqueryGetting(writer http.ResponseWriter, request *http.Request) {
	getquery := request.URL.Query().Get("getquery")
	getquery = strings.TrimSpace(getquery)
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "DBName:" + dbname + "</h4>" + `</body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "Enterd Getting Query:" + getquery + "</h4>" + `</body></html>`))
	jsonString, err, _ := getJSON(getquery)
	if err != nil {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in getting query:" + err.Error() + "</h4>" + `</body></html>`))
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + jsonString + "</h4>" + `</body></html>`))
	}

}
func executionQuery(query string) error {
	_, err := db.Exec(query)
	if err != nil {
		log.Error("error in execution query", err)
		//_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h4>" + "error in execution query:" + err.Error() + "</h4>" + `</body></html>`))
		return err
	}
	return nil
}

var dbname string

func mysqlqueryexecute(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "You Are Selected DbName:" + dbname + "</h2>" + `</body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html>
	<body>
	<h2>Enter Valid MySQL Execution Query and Execute:</h2>
	<form action="/mysqlqueryexecution" method="get">
	<label for="exequery">QUERY:</label><br>
	<input type="text" id="exequery" name="exequery" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
<h3>-----------------------------------------------------------------------------------</h3><br><br>
<h2>Enter Valid MySQL Getting Query and Execute:</h2>
<form action="/mysqlqueryGetting" method="get">
	<label for="getquery">QUERY:</label><br>
	<input type="text" id="getquery" name="getquery" value=""><br><br>
	<input type="submit" value="Submit">
	</form>
	</body>
	</html>`))
}

func postgretables(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	table := param["table"]
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ip, port, username, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	jsonString, err, count := getJSON("select * from " + table + ";")
	if err != nil {
		_, _ = writer.Write([]byte("error in getting table Data " + err.Error()))
	} else {
		_, _ = writer.Write([]byte( jsonString + "\n\n"))
		_, _ = writer.Write([]byte( "Row Count of Tables: " + strconv.Itoa(count) ))
	}
	defer db.Close()
}

func mssqltables(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname := param["dbname"]
	table := param["table"]
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		ip, username, password, port, dbname)

	db, err = sql.Open("mssql", connString)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	jsonString, err, count := getJSON("select * from " + table + ";")
	if err != nil {
		_, _ = writer.Write([]byte("error in getting table Data " + err.Error()))
	} else {
		_, _ = writer.Write([]byte( jsonString + "\n\n"))
		_, _ = writer.Write([]byte( "Row Count of Tables: " + strconv.Itoa(count) ))
	}
	defer db.Close()
}

func mysqltables(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname := param["dbname"]
	table := param["table"]
	dataSourceName := username + ":" + password + "@" + "tcp" + "(" + ip + ":" + port + ")/" + dbname
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	jsonString, err, count := getJSON("select * from " + table + ";")
	if err != nil {
		_, _ = writer.Write([]byte("error in getting table Data " + err.Error()))
	} else {
		_, _ = writer.Write([]byte( jsonString + "\n\n"))
		_, _ = writer.Write([]byte( "Row Count of Tables: " + strconv.Itoa(count) ))
	}
	defer db.Close()
}

func mssqldatabases(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname := param["dbname"]
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		ip, username, password, port, dbname)

	db, err = sql.Open("mssql", connString)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	tableName, status := getTableNames(db, "SELECT name FROM sys.tables;")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		for key, value := range tableName {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=mssqltables/" + dbname + "/" + value + ">" + strconv.Itoa(key+1) + "." + value + `</a></h3></body></html>`))
		}
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Total Number Of Tables:" + strconv.Itoa(len(tableName)) + "</h2>" + `</body></html>`))
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><h3>` + "<a href=mssqldump/" + dbname + ">" + "MSSQL-Provisionig-Deprovisioning Click Here..." + `</a></h3></body></html>`))

	}
	defer db.Close()

}

func mysqldatabases(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	dataSourceName := username + ":" + password + "@" + "tcp" + "(" + ip + ":" + port + ")/" + dbname
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	tableName, status := getTableNames(db, "SELECT table_name FROM information_schema.tables where table_schema="+"'"+dbname+"'")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		for key, value := range tableName {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=mysqltables/" + dbname + "/" + value + ">" + strconv.Itoa(key+1) + "." + value + `</a></h3></body></html>`))
		}
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Total Number Of Tables:" + strconv.Itoa(len(tableName)) + "</h2>" + `</body></html>`))

	}
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=mysqlqueryexecute/" + dbname + ">" + "Enter Here For Executing Query or Geting Query:" + `</a></h3></body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><h3>` + "<a href=mysqldump/" + dbname + ">" + "MYSQL-Provisionig-Deprovisioning Click Here..." + `</a></h3></body></html>`))

	defer db.Close()
}

func postgredatabases(writer http.ResponseWriter, request *http.Request) {
	param := mux.Vars(request)
	dbname = param["dbname"]
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ip, port, username, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	tableName, status := getTableNames(db, "SELECT table_name  FROM information_schema.tables WHERE table_schema='public'  AND table_type='BASE TABLE';")
	if status != "" {
		_, _ = writer.Write([]byte("Error getting table_name: \n" + status))
		return
	} else {
		for key, value := range tableName {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=postgretables/" + dbname + "/" + value + ">" + strconv.Itoa(key+1) + "." + value + `</a></h3></body></html>`))
		}
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Total Number Of Tables:" + strconv.Itoa(len(tableName)) + "</h2>" + `</body></html>`))
		if strings.Contains(dbname, "postgres") {
			_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><h3>` + "<a href=postgredump/" + dbname + ">" + "POSTGRE-Provisionig-Deprovisioning Click Here..." + `</a></h3></body></html>`))
		}

	}
	defer db.Close()
}

func postgresqlinput(writer http.ResponseWriter, request *http.Request) {
	ip = request.URL.Query().Get("ip")
	ip = strings.TrimSpace(ip)
	username = request.URL.Query().Get("uname")
	username = strings.TrimSpace(username)
	password = request.URL.Query().Get("password")
	password = strings.TrimSpace(password)
	port = request.URL.Query().Get("port")
	port = strings.TrimSpace(port)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ip, port, username, password, "postgres")
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
		return
	}
	err = db.Ping()
	if err != nil {
		log.Error("could not communicate eith server ", err)
		_, _ = writer.Write([]byte("could not communicate eith server " + err.Error()))
		return
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Connected SuccessFully..." + "</h2>" + `</body></html>`))
	}
	dbname, status := getDatabaseNames(db, "SELECT datname FROM pg_database;")
	if !status {
		_, _ = writer.Write([]byte("Error getting Dbnames"))
		return
	}
	result1 := postgredatabaseFraming(dbname)
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html lang="en"><head>
<style>
a:link, a:visited {
  background-color: #FFFF00;
  color: black;
  padding: 8px 4px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
}

a:hover, a:active {
  background-color: orange;
}
</style>
</head>
	<body><h4>` + result1 +
		`</h4></body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><br>` + "<h3>" + "Total Number Of DataBases: " + strconv.Itoa(len(dbname)) + "</h3>" + `</body></html>`))
	defer db.Close()
}
func postgredatabaseFraming(id []string) string {
	t := template.Must(template.New("tmpl").Parse(postgre))
	tr, _ := os.Create("databaseFraming.txt")
	err := t.Execute(tr, id)
	if err != nil {
		panic(err)
	}
	trc, _ := ioutil.ReadFile("databaseFraming.txt")
	defer tr.Close()
	return string(trc)
}
func tableFraming(id []string) string {
	t := template.Must(template.New("tmpl").Parse(tm1))
	tr, _ := os.Create("tableFraming.txt")
	err := t.Execute(tr, id)
	if err != nil {
		panic(err)
	}
	trc, _ := ioutil.ReadFile("tableFraming.txt")
	defer tr.Close()
	return string(trc)
}

const postgre = `
 <ul>
{{range $val := .}}
     <a href="postgredatabases/{{$val}}">{{$val}}</a>	 		
{{end}}
</ul>
`
const tm1 = `
 <ul>
{{range $val := .}}
     <a href="databases/tables/{{$val}}">{{$val}}</a>	 		
{{end}}
</ul>
`

func postgresql(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<center>
<h2>Enter Valid PostGreSQL Input Data:</h2>
<form action="postgreforminput">
 <label for="ip">IP/HOST:</label><br>
  <input type="text" id="ip" name="ip" value=""><br><br>
  <label for="uname">User Name:</label><br>
  <input type="text" id="uname" name="uname" value=""><br>
  <label for="password">PassWord:</label><br>
  <input type="text" id="password" name="password" value=""><br><br>
  <label for="port">Port:</label><br>
  <input type="text" id="port" name="port" value=""><br><br>
  <input type="submit" value="Submit">
</form> </center>
</body>
</html>`))
}
func mssqlforminput(writer http.ResponseWriter, request *http.Request) {
	ip = request.URL.Query().Get("ip")
	ip = strings.TrimSpace(ip)
	username = request.URL.Query().Get("uname")
	username = strings.TrimSpace(username)
	password = request.URL.Query().Get("password")
	password = strings.TrimSpace(password)
	port = request.URL.Query().Get("port")
	port = strings.TrimSpace(port)
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s",
		ip, username, password, port)

	db, err := sql.Open("mssql", connString)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	err = db.Ping()
	if err != nil {
		log.Error("could not communicate eith server ", err)
		_, _ = writer.Write([]byte("could not communicate eith server " + err.Error()))
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Connected SuccessFully..." + "</h2>" + `</body></html>`))
	}
	dbname, status := getDatabaseNames(db, "SELECT name FROM master.dbo.sysdatabases;")
	if !status {
		_, _ = writer.Write([]byte("Error getting Dbnames"))
		return
	}
	result1 := mssqldatabaseFraming(dbname)
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html lang="en"><head>
<style>
a:link, a:visited {
  background-color: #FFFF00;
  color: black;
  padding: 8px 4px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
}

a:hover, a:active {
  background-color: orange;
}
</style>
</head>
	<body><h4>` + result1 +
		`</h4></body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><br>` + "<h3>" + "Total Number Of DataBases: " + strconv.Itoa(len(dbname)) + "</h3>" + `</body></html>`))
	defer db.Close()
}

const mssqltm = `
 <ul>
{{range $val := .}}
     <a href="mssqldatabases/{{$val}}">{{$val}}</a>	 		
{{end}}
</ul>
`

func mssqldatabaseFraming(id []string) string {
	t := template.Must(template.New("tmpl").Parse(mssqltm))
	tr, _ := os.Create("databaseFraming.txt")
	err := t.Execute(tr, id)
	if err != nil {
		panic(err)
	}
	trc, _ := ioutil.ReadFile("databaseFraming.txt")
	defer tr.Close()
	return string(trc)
}
func mssql(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<center>
<h2>Enter Valid MSSQL Input Data:</h2>
<form action="mssqlforminput">
 <label for="ip">IP/HOST:</label><br>
  <input type="text" id="ip" name="ip" value=""><br><br>
  <label for="uname">User Name:</label><br>
  <input type="text" id="uname" name="uname" value=""><br>
  <label for="password">PassWord:</label><br>
  <input type="text" id="password" name="password" value=""><br><br>
  <label for="port">Port:</label><br>
  <input type="text" id="port" name="port" value=""><br><br>
  <input type="submit" value="Submit">
</form> </center>
</body>
</html>`))
}
func mysqlforminput(writer http.ResponseWriter, request *http.Request) {
	ip = request.URL.Query().Get("ip")
	ip = strings.TrimSpace(ip)
	username = request.URL.Query().Get("uname")
	username = strings.TrimSpace(username)
	password = request.URL.Query().Get("password")
	password = strings.TrimSpace(password)
	port = request.URL.Query().Get("port")
	port = strings.TrimSpace(port)
	dataSourceName := username + ":" + password + "@" + "tcp" + "(" + ip + ":" + port + ")/"

	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error("error in db connection ", err)
		_, _ = writer.Write([]byte("error in db connection " + err.Error()))
	}
	err = db.Ping()
	if err != nil {
		log.Error("could not communicate eith server ", err)
		_, _ = writer.Write([]byte("could not communicate eith server " + err.Error()))
	} else {
		_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body>` + "<h2>" + "Connected SuccessFully..." + "</h2>" + `</body></html>`))
	}
	dbname, status := getDatabaseNames(db, "SELECT schema_name FROM information_schema.schemata;")
	if !status {
		_, _ = writer.Write([]byte("Error getting Dbnames"))
		return
	}
	result1 := mysqldatabaseFraming(dbname)
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
	<html lang="en"><head>
<style>
a:link, a:visited {
  background-color: #FFFF00;
  color: black;
  padding: 8px 4px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
}

a:hover, a:active {
  background-color: orange;
}
</style>
</head>
	<body><h4>` + result1 +
		`</h4></body></html>`))
	_, _ = writer.Write([]byte(`<!DOCTYPE html><html lang="en"><body><br><br>` + "<h3>" + "Total Number Of DataBases: " + strconv.Itoa(len(dbname)) + "</h3>" + `</body></html>`))
	defer db.Close()
}
func mysqldatabaseFraming(id []string) string {
	t := template.Must(template.New("tmpl3").Parse(mysqltm))
	tr, _ := os.Create("databaseFraming.txt")
	err := t.Execute(tr, id)
	if err != nil {
		panic(err)
	}
	trc, _ := ioutil.ReadFile("databaseFraming.txt")
	defer tr.Close()
	return string(trc)
}

const mysqltm = `
 <ul>
{{range $val := .}}
     <a href="mysqldatabases/{{$val}}">{{$val}}</a>	 		
{{end}}
</ul>
`

func mysqlDb(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<center>
<h2>Enter Valid MySQL Input Data:</h2>
<form action="mysqlforminput">
 <label for="ip">IP/HOST:</label><br>
  <input type="text" id="ip" name="ip" value=""><br><br>
  <label for="uname">User Name:</label><br>
  <input type="text" id="uname" name="uname" value=""><br>
  <label for="password">PassWord:</label><br>
  <input type="text" id="password" name="password" value=""><br><br>
  <label for="port">Port:</label><br>
  <input type="text" id="port" name="port" value=""><br><br>
  <input type="submit" value="Submit">
</form> </center>
</body>
</html>`))
}

func IndexPage(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(`<!DOCTYPE html>
<html lang="en"><head>
<style>
a:link, a:visited {
  background-color: #00FFFF;
  color: black;
  padding: 8px 4px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
}

a:hover, a:active {
  background-color: green;
}
</style>
</head>
<body><br><br>` + `<div id="tfheader">

	<div class="tfclear"></div>
	</div>` + `<center><p><b><a href="mysql">MYSQL</a></b></p><center>` +
		`<center><p><b><a href="mssql">MSSQL</a></b></p><center>` +
		`<center><p><b><a href="postgresql">POSTGRESQL</a></b></p><center>` +
		`<center><p><b><a href="sqlite3">SQLITE3</a></b></p><center>` +
		`<center><p><b><a href="mongodb">MONGODB</a></b></p><center>` + `</body></html>`))
}
func getDatabaseNames(db *sql.DB, query string) ([]string, bool) {

	var databasenameArr []string
	rows, err := db.Query(query)
	if err != nil {
		log.Error("error getting tablename list", err)
		return databasenameArr, false
	}
	for rows.Next() {
		var r string
		err = rows.Scan(&r)
		if err != nil {
			log.Error("error getting tablename list")
			return databasenameArr, false
		}

		databasenameArr = append(databasenameArr, r)

	}
	defer db.Close()
	return databasenameArr, true
}
func getTableNames(db *sql.DB, query string) ([]string, string) {

	var tablenameArr []string
	rows, err := db.Query(query)
	if err != nil {
		log.Error("error getting tablename list", err)
		return tablenameArr, err.Error()
	}
	for rows.Next() {
		var r string
		err = rows.Scan(&r)
		if err != nil {
			log.Error("error getting tablename list")
			return tablenameArr, err.Error()
		}

		tablenameArr = append(tablenameArr, r)

	}

	return tablenameArr, ""
}
func getJSON(sqlString string) (string, error, int) {
	rows, err := db.Query(sqlString)
	if err != nil {
		return "", err, 0
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err, 0
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		_ = rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err := json.MarshalIndent(tableData, "", "    ")
	if err != nil {
		return "", err, 0
	}
	defer db.Close()
	return string(jsonData), nil, len(tableData)
}

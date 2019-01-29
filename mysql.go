package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sectool.go/common"

	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
)

var (
	wg sync.WaitGroup
	ch chan bool

	file         string
	f            *os.File
	command      string
	user         string
	pwd          string
	userListFile string
	pwdListFile  string
	userList     []string
	pwdList      []string

	host       string
	port       string
	timeout    int
	threads    int
	outputFile string
)

func main() {

	options := common.PublicOptions
	flag.StringVar(&user, "user", "root", "username")
	flag.StringVar(&pwd, "pwd", "", "password")
	flag.StringVar(&userListFile, "uF", "", "username file path")
	flag.StringVar(&pwdListFile, "pF", "", "password file path")
	flag.StringVar(&command, "command", "select version();", "password file path")
	flag.Parse()

	host = *options.Host
	port = *options.Port
	timeout = *options.Timeout
	threads = *options.Threads
	outputFile = *options.OutputFile
	file = *options.File

	ch = make(chan bool, threads)

	if (host == "" || port == "") && file == "" || (user == "" && userListFile == "") || (pwd == "" && pwdListFile == "") {
		flag.Usage()
		os.Exit(0)
	}

	scanList := []string{}

	if userListFile != "" {
		userList, _ = common.ReadFileLines(userListFile)
	} else if user != "" {
		userList = append(userList, user)
	}

	if pwdListFile != "" {
		pwdList, _ = common.ReadFileLines(pwdListFile)
	} else if pwd != "" {
		pwdList = append(pwdList, pwd)
	}

	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)

	if len(ipList) != 0 && len(portList) != 0 {
		for _, host := range ipList {
			for _, port := range portList {
				scanHost := fmt.Sprintf("%s:%d", host, port)
				scanList = append(scanList, scanHost)
			}
		}
	} else if file != "" {
		lines, err := common.ReadFileLines(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(portList) != 0 {
			for _, line := range lines {
				line = strings.Trim(line, " ")
				h := line
				if strings.Contains(line, ":") {
					hostPort := strings.Split(line, ":")
					h = hostPort[0]
				}
				for _, p := range portList {
					scanHost := fmt.Sprintf("%s:%d", h, p)
					scanList = append(scanList, scanHost)
				}
			}
		} else {
			for _, line := range lines {
				line = strings.Trim(line, " ")
				h := line
				p := 3306
				if strings.Contains(line, ":") {
					hostPort := strings.Split(line, ":")
					h = hostPort[0]
					p, _ = strconv.Atoi(hostPort[1])
				}

				scanHost := fmt.Sprintf("%s:%d", h, p)
				scanList = append(scanList, scanHost)
			}
		}
	}

	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		common.CheckError(err)
		defer f.Close()
	}

	fmt.Printf("host: %s\n", host)
	fmt.Printf("port: %s\n", port)
	fmt.Printf("file: %s\n", file)
	fmt.Printf("user: %s\n", user)
	fmt.Printf("pwd: %s\n", pwd)
	fmt.Printf("userFile: %s\n", userListFile)
	fmt.Printf("pwdFile: %s\n", pwdListFile)
	fmt.Printf("Number of scans: %d\n", len(scanList))

	startTime := time.Now()
	for _, line := range scanList {
		ch <- true
		wg.Add(1)

		pair := strings.SplitN(line, ":", 2)
		host := pair[0]
		port, _ := strconv.Atoi(pair[1])
		go scan(host, port)
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)
}

func scan(ip string, port int) {
	defer func() {
		<-ch
		wg.Done()
	}()

	for _, username := range userList {
		for _, password := range pwdList {
			if isLogin, client, err := mysqlConnect(ip, port, username, password); isLogin {

				var line = fmt.Sprintf("%s:%d %s %s\n", ip, port, username, password)
				f.WriteString(line)
				fmt.Printf("\033[0;32m[log] %s:%d %s %s\033[0m\n", ip, port, username, password)

				output, err := mysqlExec(client, command)
				if err != nil {
					fmt.Println("Failed to exec: " + err.Error())
				}
				fmt.Printf("\033[0;32m[sql] %s:%d \n%s\033[0m", ip, port, output)
				return
			} else {
				// fmt.Printf("[err] %s:%d %s %s\n", ip, port, username, password)
				fmt.Println(ip, port, err)
				// if strings.HasPrefix(err.Error(), "Error ") {
				// 	errMsg := strings.SplitN(err.Error(), ":", 2)
				// 	fmt.Printf("[err] %s\n", strings.Trim(errMsg[1], " "))
				// }
			}
		}
	}
}

func mysqlConnect(ip string, port int, username string, password string) (bool, *sql.DB, error) {

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&timeout=%ds", username, password, ip, port, timeout)
	db, err := sql.Open("mysql", DSN)

	if err != nil {
		return false, nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		// fmt.Println("Connect Error", err)
		return false, nil, err
	}

	return true, db, nil
}

func mysqlExec(db *sql.DB, sqlStr string) (string, error) {

	defer db.Close()

	// Execute the query
	rows, err := db.Query(sqlStr)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader(columns)
	// table.SetCaption(true, "Movie ratings.")

	var data [][]string
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Now do something with the data.
		// Here we just print each column as a string.

		var v []string
		var value string
		for _, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			v = append(v, value)
		}
		data = append(data, v)
	}

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
	return buf.String(), nil

	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return "", err
}

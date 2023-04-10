package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/Unknwon/goconfig"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wonderivan/logger"
)

// 定义全局变量
var (
	// 数据库连接串
	DBConn string
	// 数据库最大连接数
	MaxConn int
	// 源地址根路径
	SrcRoot string
	// 目标地址跟路径
	DestRoot string
	// 源地址code
	SrcCode int
	// 目标地址code
	DestCode int
	//设置最大线程数
	MaxThreads int
	//最大任务数
	MaxTasks int
	// 空闲磁盘大小
	DiskSize int
	// 检查磁盘
	CheckDisk string
	// 程序开始执行时间
	StartHour int
	// 程序结束执行时间
	EndHour int
	// 数据通道
	dataChan chan int64
)
var db *sql.DB

func Worke(key int64) {
	// 获取dicom和JPG文件路径
	dicom, jpg := getFilePath(key)
	if dicom != "" {
		dicomsrcpath := SrcRoot + dicom
		logger.Debug(key, ": dicom源文件路径：", dicomsrcpath)
		dicomdestpath := DestRoot + dicom
		logger.Debug(key, ": dicom目标文件路径： ", dicomdestpath)
		// 判断路径文件夹是否存在，不存在，创建文件夹
		_, err := copyFile(dicomsrcpath, dicomdestpath)
		if err != nil {
			logger.Info("Dicom文件不存在", dicomsrcpath)
			updateFileFlag(key)
		} else {
			logger.Info("Dicom拷贝完成: ", dicomsrcpath)
			// 更新instance表中的localtion_code
			updateLocationCode(key)
		}
	}
	if jpg != "" {
		jpgsrcpath := SrcRoot + jpg
		logger.Debug(key, ": jpg源文件路径：", jpgsrcpath)
		jpgdestpath := DestRoot + jpg
		logger.Debug(key, ": jpg目标文件路径： ", jpgdestpath)
		_, err := copyFile(jpgsrcpath, jpgdestpath)
		check(err)
		//logger.Info("Jpg拷贝完成: ", jpgsrcpath)
	}
	// 删除任务列表中数据
	// deleteData(d.key)
}

func main() {
	// 获取可执行文件的路径
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir)
	// 配置log
	logger.SetLogger(dir + "/log.json")
	logger.Debug("当前可执行文件路径：" + dir)
	// 读取配置文件
	readConfigFile(dir)
	// instancekey数据通道
	dataChan = make(chan int64)
	//dataChan = make(chan int64)
	// 	//池子
	ch := make(chan struct{}, MaxThreads)
	// 初始化数据库
	initDB()
	go func() { //这是一个独立的协程 保证可以接受到每个用户的请求
		for {
			select {
			case instance_key := <-dataChan:
				ch <- struct{}{}
				go func(key int64) {
					defer func() {
						<-ch
					}()
					Worke(instance_key)
				}(instance_key)
			}
		}
	}()
	// 主程序逻辑
	for {
		// if db.Ping() != nil {
		// 	initDB()
		// }
		// 1.检查是否时程序运行时间段
		// if WorkTime() {
		// 2.获取磁盘剩余空间
		size := int(GetDiskSize(CheckDisk))
		logger.Info("目标磁盘剩余空间：", size, "GB")
		if size < DiskSize {
			logger.Info("目标磁盘剩余空间达到设定值，程序退出")
			os.Exit(1)
		}
		// 3.获取instance_key
		getinstancekey()
		// }
	}
}

// io.copy()来复制
// 参数说明：
// src: 源文件路径
// dest: 目标文件路径
// key :值不为空是更新instance表中的localtion_code值
func copyFile(src, dest string) (int64, error) {
	if !Exist(src) {
		logger.Debug(src, "文件不存在.....")
		return 0, errors.New("文件不存在")
	} else {
		// 判断路径文件夹是否存在，不存在，创建文件夹
		CheckPath(dest)
		logger.Info("开始拷贝文件：", src)
		file1, err := os.Open(src)
		if err != nil {
			return 0, err
		}
		defer file1.Close()
		file2, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return 0, err
		}
		defer file2.Close()
		return io.Copy(file2, file1)
	}
}

//因为要多次检查错误，所以建立一个函数。
func check(err error) {
	if err != nil {
		logger.Debug(err)
	}
}

//初始化全局变量
func readConfigFile(dir string) {
	logger.Debug("开始读取配置文件....")
	cfg, err := goconfig.LoadConfigFile(dir + "/config.ini")
	if err != nil {
		logger.Debug("无法加载配置文件：%s", err)
	}
	DBConn, _ = cfg.GetValue("Mysql", "DBConn")
	MaxConn, _ = cfg.Int("Mysql", "MaxConn")
	SrcRoot, _ = cfg.GetValue("General", "SrcRoot")
	DestRoot, _ = cfg.GetValue("General", "DestRoot")
	SrcCode, _ = cfg.Int("General", "SrcCode")
	DestCode, _ = cfg.Int("General", "DestCode")
	MaxThreads, _ = cfg.Int("General", "MaxThreads")
	MaxTasks, _ = cfg.Int("General", "MaxTasks")
	DiskSize, _ = cfg.Int("General", "DiskSize")
	CheckDisk, _ = cfg.GetValue("General", "CheckDisk")
	StartHour, _ = cfg.Int("General", "StartHour")
	EndHour, _ = cfg.Int("General", "EndHour")

	logger.Debug("DBConn: ", DBConn)
	logger.Debug("MaxConn: ", MaxConn)
	logger.Debug("SrcRoot: ", SrcRoot)
	logger.Debug("DestRoot: ", DestRoot)
	logger.Debug("SrcCode: ", SrcCode)
	logger.Debug("DestCode: ", DestCode)
	logger.Debug("MaxThreads: ", MaxThreads)
	logger.Debug("MaxTasks: ", MaxTasks)
	logger.Debug("DiskSize: ", DiskSize)
	logger.Debug("CheckDisk: ", CheckDisk)
	logger.Debug("StartHour: ", StartHour)
	logger.Debug("EndHour: ", EndHour)
}

func getFilePath(instancekey int64) (string, string) {
	logger.Debug("开始获取文件路径通过key:", instancekey)
	logger.Debug("***************************数据库空闲连接数：", db.Stats())
	sql := "SELECT i.file_name,m.img_file_name FROM instance i LEFT JOIN image m ON i.instance_key=m.instance_key WHERE i.instance_key=?;"
	var dicomPath string
	var jpgPath string
	row := db.QueryRow(sql, instancekey)
	// 1.获取dicom 和jpg 路径
	err := row.Scan(&dicomPath, &jpgPath)
	check(err)
	logger.Debug(instancekey, "获取到的文件名是：", dicomPath)
	logger.Debug(instancekey, "获取到的文件名是：", jpgPath)
	return dicomPath, jpgPath
}

func updateLocationCode(instancekey int64) {
	logger.Debug("开始更新instance表中的location_code")
	sql := "UPDATE instance set location_code = ? where instance_key=?;"
	db.Exec(sql, DestCode, instancekey)
}

func updateFileFlag(instancekey int64) {
	logger.Debug("开始更新intance文件不存在标志：", instancekey)
	sql := "UPDATE instance set FileExist=-1 where instance_key= ?;"
	db.Exec(sql, instancekey)
}

func GetDiskSize(disk string) (size int64) {
	logger.Debug("获取磁盘大小......")
	kernel := syscall.MustLoadDLL("kernel32.dll")
	proc := kernel.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := int64(0)
	proc.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(disk))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)))
	size = lpFreeBytesAvailable / 1024 / 1024 / 1024.0 //GB
	return size
}

func WorkTime() bool {
	// 当前时间
	currentHour := time.Now().Hour()
	if StartHour <= currentHour && currentHour <= EndHour {
		return true
	}
	logger.Debug("不是工作时间,等待工作时间.....")
	return false
}

func CheckPath(path string) {
	dir, _ := filepath.Split(path)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		}
	}
}

func getinstancekey() {
	// sql := `Select ins.instance_key from instance ins
	// left join instance_task_list itl on ins.instance_key = itl.instance_key
	// where itl.instance_key is null and location_code = ? and FileExist != -1
	// order by instance_key ASC limit ?;`
	sql := `Select ins.instance_key from instance ins 
	where location_code = ? and FileExist != -1 
	order by instance_key ASC limit ?;`
	logger.Info("***************************数据库空闲连接数：", db.Stats())
	rows, err := db.Query(sql, SrcCode, MaxTasks)
	if err != nil {
		logger.Fatal(err)
		return
	} else {
		for rows.Next() {
			var instance_key int64
			err = rows.Scan(&instance_key)
			check(err)
			logger.Debug("获取到的instance_key是： ", instance_key)
			dataChan <- instance_key
			//insertData(instance_key)
		}
		rows.Close()
	}
}

func initDB() {
	db, _ = sql.Open("mysql", DBConn)
	// 数据库最大连接数
	db.SetMaxOpenConns(MaxConn)
	//db.SetMaxIdleConns(MaxTasks)
	err := db.Ping()
	if err != nil {
		panic(err.Error())
	}
	logger.Debug("数据库连接成功...")
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

package v1

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

//任务
type Job interface {
	// do something...
	Do()
}

//worker 工人
type Worker struct {
	JobQueue chan Job  //任务队列
	Quit     chan bool //停止当前任务
}

//新建一个 worker 通道实例  新建一个工人
func NewWorker() Worker {
	return Worker{
		JobQueue: make(chan Job), //初始化工作队列为null
		Quit:     make(chan bool),
	}
}

/*
整个过程中 每个Worker(工人)都会被运行在一个协程中，
在整个WorkerPool(领导)中就会有num个可空闲的Worker(工人)，
当来一条数据的时候，领导就会小组中取一个空闲的Worker(工人)去执行该Job，
当工作池中没有可用的worker(工人)时，就会阻塞等待一个空闲的worker(工人)。
每读到一个通道参数 运行一个 worker
*/

func (w Worker) Run(wq chan chan Job) {
	//这是一个独立的协程 循环读取通道内的数据，
	//保证 每读到一个通道参数就 去做这件事，没读到就阻塞
	go func() {
		for {
			wq <- w.JobQueue //注册工作通道  到 线程池
			select {
			case job := <-w.JobQueue: //读到参数
				job.Do()
			case <-w.Quit: //终止当前任务
				return
			}
		}
	}()
}

//workerpool 领导
type WorkerPool struct {
	workerlen   int      //线程池中  worker(工人) 的数量
	JobQueue    chan Job //线程池的  job 通道
	WorkerQueue chan chan Job
}

func NewWorkerPool(workerlen int) *WorkerPool {
	return &WorkerPool{
		workerlen:   workerlen,                      //开始建立 workerlen 个worker(工人)协程
		JobQueue:    make(chan Job),                 //工作队列 通道
		WorkerQueue: make(chan chan Job, workerlen), //最大通道参数设为 最大协程数 workerlen 工人的数量最大值
	}
}

//运行线程池
func (wp *WorkerPool) Run() {
	//初始化时会按照传入的num，启动num个后台协程，然后循环读取Job通道里面的数据，
	//读到一个数据时，再获取一个可用的Worker，并将Job对象传递到该Worker的chan通道
	logger.Debug("初始化worker")
	for i := 0; i < wp.workerlen; i++ {
		//新建 workerlen worker(工人) 协程(并发执行)，每个协程可处理一个请求
		worker := NewWorker() //运行一个协程 将线程池 通道的参数  传递到 worker协程的通道中 进而处理这个请求
		worker.Run(wp.WorkerQueue)
	}

	// 循环获取可用的worker,往worker中写job
	go func() { //这是一个单独的协程 只负责保证 不断获取可用的worker
		for {
			select {
			case job := <-wp.JobQueue: //读取任务
				//尝试获取一个可用的worker作业通道。
				//这将阻塞，直到一个worker空闲
				worker := <-wp.WorkerQueue
				worker <- job //将任务 分配给该工人
			}
		}
	}()
}

//----------------------------------------------
type Dosomething struct {
	key int64
}

func (d *Dosomething) Do() {
	// 获取dicom和JPG文件路径
	// 增加时间间隔，减少任务执行频率
	time.Sleep(3 * time.Second)
	dicom, jpg := getFilePath(d.key)
	if dicom != "" {
		dicomsrcpath := SrcRoot + dicom
		logger.Debug(d.key, ": dicom源文件路径：", dicomsrcpath)
		dicomdestpath := DestRoot + dicom
		logger.Debug(d.key, ": dicom目标文件路径： ", dicomdestpath)
		// 判断路径文件夹是否存在，不存在，创建文件夹
		_, err := copyFile(dicomsrcpath, dicomdestpath)
		if err != nil {
			logger.Debug(d.key, "的dicom文件不存在")
			updateFileFlag(d.key)
		} else {
			logger.Info(d.key, "的dicom拷贝完成...")
			// 更新instance表中的localtion_code
			updateLocationCode(d.key)
		}
	}
	if jpg != "" {
		jpgsrcpath := SrcRoot + jpg
		logger.Debug(d.key, ": jpg源文件路径：", jpgsrcpath)
		jpgdestpath := DestRoot + jpg
		logger.Debug(d.key, ": jpg目标文件路径： ", jpgdestpath)
		_, err := copyFile(jpgsrcpath, jpgdestpath)
		check(err)
		logger.Info(d.key, "的jpg拷贝完成")
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
	// 初始化数据库
	initDB()
	// 注册工作池，传入任务
	// 参数1 初始化worker(工人)设置最大线程数
	wokerPool := NewWorkerPool(MaxThreads)
	wokerPool.Run() //有任务就去做，没有就阻塞，任务做不过来也阻塞
	// 处理任务：
	go func() { //这是一个独立的协程 保证可以接受到每个用户的请求
		for {
			select {
			case instance_key := <-dataChan:
				sc := &Dosomething{key: instance_key}
				wokerPool.JobQueue <- sc //往线程池 的通道中 写参数   每个参数相当于一个请求  来了100万个请求
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

func insertData(instancekey int64) {
	logger.Debug("开始插入任务表：", instancekey)
	sql := "insert INTO instance_task_list set instance_key = ?"
	db.Exec(sql, instancekey)
}

func deleteData(instancekey int64) {
	logger.Debug("开始插入任务表：", instancekey)
	sql := "delete from instance_task_list where instance_key = ?"
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

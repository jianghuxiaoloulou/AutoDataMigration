package main

import (
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestNewWorker(t *testing.T) {
	tests := []struct {
		name string
		want Worker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWorker(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWorker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorker_Run(t *testing.T) {
	type args struct {
		wq chan chan Job
	}
	tests := []struct {
		name string
		w    Worker
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.w.Run(tt.args.wq)
		})
	}
}

func TestNewWorkerPool(t *testing.T) {
	type args struct {
		workerlen int
	}
	tests := []struct {
		name string
		args args
		want *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWorkerPool(tt.args.workerlen); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWorkerPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkerPool_Run(t *testing.T) {
	tests := []struct {
		name string
		wp   *WorkerPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wp.Run()
		})
	}
}

func TestDosomething_Do(t *testing.T) {
	tests := []struct {
		name string
		d    *Dosomething
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.Do()
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_copyFile(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := copyFile(tt.args.src, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("copyFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_check(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(tt.args.err)
		})
	}
}

func Test_readConfigFile(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readConfigFile(tt.args.dir)
		})
	}
}

func Test_getFilePath(t *testing.T) {
	type args struct {
		instancekey int64
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getFilePath(tt.args.instancekey)
			if got != tt.want {
				t.Errorf("getFilePath() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getFilePath() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_updateLocationCode(t *testing.T) {
	type args struct {
		instancekey int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateLocationCode(tt.args.instancekey)
		})
	}
}

func Test_updateFileFlag(t *testing.T) {
	type args struct {
		instancekey int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateFileFlag(tt.args.instancekey)
		})
	}
}

func Test_insertData(t *testing.T) {
	type args struct {
		instancekey int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//insertData(tt.args.instancekey)
		})
	}
}

func Test_deleteData(t *testing.T) {
	type args struct {
		instancekey int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//deleteData(tt.args.instancekey)
		})
	}
}

func TestGetDiskSize(t *testing.T) {
	type args struct {
		disk string
	}
	tests := []struct {
		name     string
		args     args
		wantSize int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSize := GetDiskSize(tt.args.disk); gotSize != tt.wantSize {
				t.Errorf("GetDiskSize() = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func TestWorkTime(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WorkTime(); got != tt.want {
				t.Errorf("WorkTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckPath(tt.args.path)
		})
	}
}

func Test_getinstancekey(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getinstancekey()
		})
	}
}

func Test_initDB(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initDB()
		})
	}
}

func TestExist(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exist(tt.args.filename); got != tt.want {
				t.Errorf("Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

package main

import (
	"NAS-Server-Web/DatabaseService"
	"NAS-Server-Web/shared/configurations"
	"errors"
	"os"
	"path"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		println("must give Username, Password and MemoryAllocated (GB)")
		os.Exit(1)
	}

	err := configurations.UpdateConfigurations()
	if err != nil {
		panic(err)
	}

	db, err := DatabaseService.NewDatabaseService()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	memoryBits, err := strconv.Atoi(os.Args[3])
	memoryMegaBytes := memoryBits * 1024 * 1024 * 1024

	if err := db.AddUser(os.Args[1], os.Args[2], memoryMegaBytes); err != nil {
		println("cannot add user ", err.Error())
		os.Exit(1)
	}

	err = os.Mkdir(configurations.GetBaseFilesPath(), os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	fullPath := path.Join(configurations.GetBaseFilesPath(), os.Args[1])
	_ = os.Mkdir(fullPath, os.ModePerm)
}

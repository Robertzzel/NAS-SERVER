package main

import (
	"NAS-Server-Web/services"
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

	configs, err := services.NewConfigsService()
	if err != nil {
		panic(err)
	}

	db, err := services.NewDatabaseService()
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

	err = os.Mkdir(configs.BaseFilesBath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	fullPath := path.Join(configs.BaseFilesBath, os.Args[1])
	os.Mkdir(fullPath, os.ModePerm)
}

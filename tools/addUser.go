package main

import (
	. "NAS-Server-Web/shared/Services"
	"NAS-Server-Web/shared/configurations"
	"errors"
	"log"
	"os"
	"path"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		println("must give Username, Password and MemoryAllocated (GB)")
		os.Exit(1)
	}

	log.Println("Updateing configurations...")
	err := configurations.UpdateConfigurations()
	if err != nil {
		panic(err)
	}

	log.Println("Create a instance of Database configurations...")
	db, err := NewDatabaseService()
	if err != nil {
		panic(err)
	}

	memoryBits, err := strconv.Atoi(os.Args[3])
	memoryMegaBytes := memoryBits * 1024 * 1024 * 1024

	log.Println("Adding user...")
	if added := db.AddUser(os.Args[1], os.Args[2], memoryMegaBytes); !added {
		println("cannot add user ", err.Error())
		os.Exit(1)
	}

	log.Println("Creating file structure...")
	err = os.Mkdir(configurations.GetBaseFilesPath(), os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	fullPath := path.Join(configurations.GetBaseFilesPath(), os.Args[1])
	_ = os.Mkdir(fullPath, os.ModePerm)

	log.Println("Finished.")
}

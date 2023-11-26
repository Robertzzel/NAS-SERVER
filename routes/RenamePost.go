package routes

import (
	"NAS-Server-Web/services/filesService"
	"NAS-Server-Web/services/sessionService"
	"encoding/json"
	"net/http"
	"path/filepath"
)

func RenamePost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("ftp")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := sessionService.GetSession(cookie)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	newName, hasNewName := data["new-name"]
	oldPath, hasOldPath := data["old-path"]
	if !hasOldPath || !hasNewName {
		return
	}

	newName = filepath.Clean(newName)
	oldPath = filepath.Clean(oldPath)

	newName = filepath.Join(filepath.Dir(oldPath), newName)
	fileDirectory := filepath.Dir(oldPath)
	if fileDirectory == "." || fileDirectory == "/" {
		fileDirectory = ""
	}

	fullOldPath := filepath.Join(session.BasePath, oldPath)
	fullNewPath := filepath.Join(session.BasePath, newName)
	if err = filesService.RenameFile(fullOldPath, fullNewPath); err != nil {
		http.Redirect(w, r, "/home/"+fileDirectory, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/home/"+fileDirectory, http.StatusSeeOther)
}

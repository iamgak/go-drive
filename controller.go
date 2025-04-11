package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/iamgak/go-drive/models"
	"github.com/iamgak/go-drive/pkg"
)

type FileEntry struct {
	Name string
	Path string
	Icon string
}

type DriveTemplateData struct {
	CurrentPath string
	ParentPath  string
	ShowBack    bool
	Entries     []FileEntry
}

func (app *Application) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login",
	})
}

func (app *Application) ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title": "Register",
	})
}

func (app *Application) UserActivateAccount(c *gin.Context) {
	token := c.Param("token")
	err := app.Model.UsersORM.ActivateAccount(token)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrNoRecord {
			app.ErrorJSONResponse(c.Writer, http.StatusNotFound, err.Error())
			return
		}

		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	app.sendJSONResponse(c.Writer, http.StatusOK, "Account Activated Successfully")
}

func (app *Application) UserLogin(c *gin.Context) {
	var creds *models.UserStruct
	if err := c.ShouldBindJSON(&creds); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.UsersORM.ValidateUserData(creds, false)
	if len(validator.Errors) != 0 {
		c.JSON(http.StatusBadRequest, validator)
		return
	}

	token, err := app.Model.UsersORM.LoginUser(c.Request.Context(), creds)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrAccountInActive {
			app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, err.Error())
			return
		}

		if err == pkg.ErrInvalidCredentials {
			app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, err.Error())
			return
		}

		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "ldata",
		Value:    token,
		HttpOnly: true,
		// Secure:   true, // Only for HTTPS
		Path:     "/",
		MaxAge:   4 * 3600, // 4 hour
		SameSite: http.SameSiteStrictMode,
	})
	app.sendJSONResponse(c.Writer, http.StatusOK, "Login Successfull")
}

func (app *Application) UserRegister(c *gin.Context) {
	var creds *models.UserStruct
	if err := c.ShouldBindJSON(&creds); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.UsersORM.ValidateUserData(creds, true)
	if len(validator.Errors) != 0 {
		c.JSON(http.StatusBadRequest, validator)
		return
	}

	if err := app.Model.UsersORM.RegisterUser(c.Request.Context(), creds.Email, creds.Password, c.ClientIP()); err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}

	app.sendJSONResponse(c.Writer, http.StatusCreated, "Registration Successfully")
}

func (app *Application) CreateFolder(c *gin.Context) {
	type Req struct {
		SavePath   string `json:"save_path"`
		FolderName string `json:"folder_name"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil || req.FolderName == "" {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Missing save_path or folder_name")
		return
	}

	// Combine base directory + save path + folder name
	relPath := filepath.Join(req.SavePath, req.FolderName)
	fullPath, err := filepath.Abs(filepath.Join(app.BaseDir, relPath))
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to resolve path")
		return
	}

	baseAbs, err := filepath.Abs(app.BaseDir)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to resolve base path")
		return
	}

	if !strings.HasPrefix(fullPath, baseAbs) {
		app.ErrorJSONResponse(c.Writer, http.StatusForbidden, "Access denied")
		return
	}

	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Could not create folder")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: fmt.Sprintf("Folder Created: %s ", relPath)}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		log.Println("Error creating folder activity ", err)
	}
	app.sendJSONResponse(c.Writer, http.StatusOK, "Folder created")
}

func (app *Application) DeleteFileOrFolder(c *gin.Context) {
	type Req struct {
		Path string `json:"path"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Invalid input")
		return
	}

	target := filepath.Clean(filepath.Join(app.BaseDir, req.Path))
	if !strings.HasPrefix(target, app.BaseDir) {
		app.ErrorJSONResponse(c.Writer, http.StatusForbidden, "Access denied")
		return
	}

	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		app.ErrorJSONResponse(c.Writer, http.StatusNotFound, "Path not found")
		return
	}
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to access path")
		return
	}

	if info.IsDir() {
		err = os.RemoveAll(target)
	} else {
		err = os.Remove(target)
	}
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to delete")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: fmt.Sprintf("File Deleted: %s ", target)}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		log.Println("Error deleting file activity ", err)
	}
	app.sendJSONResponse(c.Writer, http.StatusOK, "Folder deleted")
}

func (app *Application) RenameFolder(c *gin.Context) {
	type Req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPath == "" {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Invalid input")
		return
	}

	oldFull := filepath.Clean(filepath.Join(app.BaseDir, req.OldPath))
	newFull := filepath.Clean(filepath.Join(app.BaseDir, req.NewPath))

	if !strings.HasPrefix(oldFull, app.BaseDir) || !strings.HasPrefix(newFull, app.BaseDir) {
		app.ErrorJSONResponse(c.Writer, http.StatusForbidden, "Access denied")
		return
	}

	err := os.Rename(oldFull, newFull)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Could not rename folder")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: fmt.Sprintf("File Renamed: %s to %s ", req.OldPath, req.NewPath)}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		log.Println("Error renaming file activity ", err)
	}
	app.sendJSONResponse(c.Writer, http.StatusOK, "Folder Renamed")
}

func (app *Application) UploadFile(c *gin.Context) {

	uploadDir := filepath.Join(app.BaseDir, c.PostForm("save_path")) // e.g., /drive/6/new

	// Ensure directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to create upload directory: "+err.Error())
		return
	}

	if err := c.Request.ParseMultipartForm(3 << 20); err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "No file found in request: "+err.Error())
		return
	}
	defer file.Close()

	// Validate size
	if header.Size > 2*1024*1024 {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "File size exceeds 2MB limit")
		return
	}

	// Read first 512 bytes to detect MIME
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to read file header")
		return
	}
	contentType := http.DetectContentType(buffer)

	// Reset read pointer
	file.Seek(0, io.SeekStart)

	// Allowed types
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
	if !allowedTypes[contentType] {
		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Invalid file type: "+contentType)
		return
	}

	// Create destination file
	dstPath := filepath.Join(uploadDir, header.Filename)

	out, err := os.Create(dstPath)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to create file: "+err.Error())
		return
	}
	defer out.Close()

	// Save the file
	_, err = io.Copy(out, file)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to save file")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: "File Uploaded: " + header.Filename}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		log.Println("Error saving activity ", err)
	}
	app.sendJSONResponse(c.Writer, http.StatusOK, "File uploaded successfully")
}

func (app *Application) DriveListing(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	fullPath := filepath.Clean(filepath.Join(app.BaseDir, path))

	baseAbs, err := filepath.Abs(app.BaseDir)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to resolve base path")
		return
	}

	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Failed to resolve path")
		return
	}

	if _, err := os.Stat(app.BaseDir); os.IsNotExist(err) {
		err := os.MkdirAll(app.BaseDir, 0755) // initial folder for user where he roam
		if err != nil {
			app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Base directory not found and could not be created")
			return
		}
	}

	if !strings.HasPrefix(fullAbs, baseAbs) {
		app.ErrorJSONResponse(c.Writer, http.StatusForbidden, "Access denied")
		return
	}

	info, err := os.Stat(fullAbs)
	if os.IsNotExist(err) {
		app.ErrorJSONResponse(c.Writer, http.StatusNotFound, "File or directory not found")
		return
	}
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if info.IsDir() {
		files, err := os.ReadDir(fullAbs)
		if err != nil {
			app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Unable to read directory")
			return
		}

		var entries []FileEntry
		for _, f := range files {
			entry := FileEntry{
				Name: f.Name(),
				Path: filepath.Join(path, f.Name()),
				Icon: "📁",
			}
			if !f.IsDir() {
				entry.Icon = "📄"
			}
			entries = append(entries, entry)
		}

		data := DriveTemplateData{
			CurrentPath: path,
			ParentPath:  filepath.Dir(path),
			ShowBack:    path != "",
			Entries:     entries,
		}

		tmpl, err := template.ParseFiles("templates/drive.html")
		if err != nil {
			app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Template error")
			return
		}
		tmpl.Execute(c.Writer, data)
		return
	}

	data, err := os.ReadFile(fullAbs)
	if err != nil {
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Unable to read file")
		return
	}
	mimeType := mime.TypeByExtension(filepath.Ext(fullAbs))
	if mimeType == "" {
		mimeType = http.DetectContentType(data[:512])
	}
	c.Data(http.StatusOK, mimeType, data)
}

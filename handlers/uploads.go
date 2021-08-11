package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func registerPublicRoutes(g *echo.Group) {

	//this needs to be in a private folder - not in public domain, who will delete the files if its only uploaded an never used
	g.POST("/upload", FileUpload)

}

func returnResponse(c echo.Context, i int, s string) error {
	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"Invalid params : please provide a excel file (.xlsx)"})
}

func FileUpload(c echo.Context) error {

	file, err := c.FormFile("file")

	if err != nil {
		//http: no such file
		// errMess := fmt.Sprintf("Invalid params : please provide a excel file (.xlsx)")
		// fmt.Println(errMess)
		return returnResponse(c, http.StatusBadRequest, "invalid params : please provide a excel file (.xlsx)")
	}
	fileType := strings.Split(file.Filename, ".")
	fmt.Println(fileType)
	if len(fileType) < 2 || fileType[1] != "xlsx" {
		// errMess := fmt.Sprintf("Invalid file foramt. Please upload an excel file with .xlsx")
		// fmt.Println(errMess)
		return returnResponse(c, http.StatusBadRequest, "invalid params : please provide a excel file (.xlsx)")
	}
	src, err := file.Open()
	if err != nil {
		return returnResponse(c, http.StatusBadRequest, "invalid params : please provide a excel file (.xlsx)")
	}
	defer src.Close()

	tmpfile, err := ioutil.TempFile("/tmp", fileType[0]+"*"+".xlsx")
	if err != nil {
		log.Panic(err)
		return returnResponse(c, http.StatusBadRequest, "Error uploading file")
	}
	retFileName := strings.Split(tmpfile.Name(), "/")[2]

	// defer os.Remove(tmpfile.Name()) // clean up

	fileBytes, err := ioutil.ReadAll(src)
	if err != nil {
		log.Panic(err)
		return returnResponse(c, http.StatusBadRequest, "Error uploading file")
	}
	if _, err := tmpfile.Write(fileBytes); err != nil {
		log.Panic(err)
		return returnResponse(c, http.StatusBadRequest, "Error uploading file")
	}
	if err := tmpfile.Close(); err != nil {
		log.Panic(err)
		return returnResponse(c, http.StatusBadRequest, "Error uploading file")
	}

	return c.JSON(http.StatusOK, struct {
		M string `json:"message"`
		F string `json:"file_name"`
	}{"Success", retFileName})
}

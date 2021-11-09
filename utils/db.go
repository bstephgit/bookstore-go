package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Db struct {
	Connected        bool
	DbObj            *sql.DB
	ConnectionString string
}

type Subject struct {
	Name    string
	Id      uint
	NbBooks uint
}

type BookLine struct {
	Id    uint
	Title string
}

type Book struct {
	Id          int
	Title       string
	Authors     string
	Year        int
	Description string
}

type BookDownload struct {
	BookId     int
	StorageId  int
	FileSize   int
	FileId     string
	FileName   string
	Vendor     string
	VendorCode string
}

type Database struct {
	Host     string
	User     string
	Password string
}

var DatabaseObj Db

func init() {
	//fmt.Println("Init utils/Db package")
	DatabaseObj.Connected = false
	DatabaseObj.DbObj = nil
	DatabaseObj.ConnectionString = ""
}

func GetDatabaseObject() *Db {
	return &DatabaseObj
}

func DbConnect(config Database) error {

	if !DatabaseObj.Connected {

		passwd := config.Password
		connectionstr := config.User + ":" + string(passwd) + "@tcp(" + config.Host + ")/booksdb"

		//fmt.Println(connectionstr)
		db, err := sql.Open("mysql", connectionstr)

		if err == nil {
			DatabaseObj.DbObj = db
			DatabaseObj.Connected = true
			DatabaseObj.ConnectionString = connectionstr
			DatabaseObj.DbObj.SetConnMaxLifetime(time.Minute * 3)
			fmt.Println("Connected to database")
		} else {
			DatabaseObj.DbObj = nil
			DatabaseObj.Connected = false
		}
		return err
	}

	return errors.New("Already connected")
}

func GetSubjects() ([]Subject, error) {

	if DatabaseObj.Connected {
		rows, err := DatabaseObj.DbObj.Query("SELECT ID,NAME,COUNT(ID) FROM IT_SUBJECT S, BOOKS_SUBJECTS_ASSOC WHERE SUBJECT_ID=S.ID GROUP BY SUBJECT_ID")

		if err != nil {
			return nil, err
		}
		var subjects []Subject
		for rows.Next() {
			var sub Subject
			rows.Scan(&sub.Id, &sub.Name, &sub.NbBooks)
			subjects = append(subjects, sub)
		}
		return subjects, nil
	}
	return nil, errors.New("Not Connected")
}

func GetSubjectBooks(subjectID int) ([]BookLine, error) {

	if DatabaseObj.Connected {
		query := fmt.Sprintf("SELECT B.ID,B.TITLE FROM BOOKS B, BOOKS_SUBJECTS_ASSOC WHERE SUBJECT_ID=%d AND BOOK_ID=B.ID", subjectID)
		rows, err := DatabaseObj.DbObj.Query(query)

		if err != nil {
			return nil, err
		}

		var books []BookLine
		for rows.Next() {
			var bookLine BookLine
			rows.Scan(&bookLine.Id, &bookLine.Title)
			books = append(books, bookLine)
		}
		return books, nil
	}
	return nil, errors.New("Not Connected")
}

func GetBook(id int) (*Book, error) {
	if DatabaseObj.Connected {
		query := fmt.Sprintf("SELECT ID,TITLE,AUTHORS,YEAR, DESCR FROM BOOKS WHERE ID=%d", id)
		row, err := DatabaseObj.DbObj.Query(query)

		if err != nil {
			return nil, err
		}
		book := &Book{}
		if row.Next() {
			row.Scan(&book.Id, &book.Title, &book.Authors, &book.Year, &book.Description)
		} else {
			return nil, errors.New("Book not found")
		}
		return book, nil
	}
	return nil, errors.New("Not Connected")
}

func GetDownloadInfo(bookid int) (*BookDownload, error) {
	if DatabaseObj.Connected {

		//query := fmt.Sprintf("SELECT BOOK_ID,STORE_ID,FILE_ID,FILE_SIZE,FILE_NAME,VENDOR,VENDOR_CODE FROM BOOKS_LINKS,FILE_STORE WHERE BOOK_ID=%d", bookid)
		query := fmt.Sprintf("SELECT BOOK_ID,STORE_ID,FILE_ID,FILE_SIZE,FILE_NAME,VENDOR,VENDOR_CODE FROM BOOKS_LINKS, FILE_STORE FS WHERE BOOK_ID=%d AND FS.ID=STORE_ID", bookid)
		row, err := DatabaseObj.DbObj.Query(query)

		if err != nil {
			return nil, err
		}

		if row.Next() {
			book_dl := &BookDownload{}

			row.Scan(&book_dl.BookId, &book_dl.StorageId, &book_dl.FileId, &book_dl.FileSize, &book_dl.FileName, &book_dl.Vendor, &book_dl.VendorCode)
			return book_dl, nil
		} else {
			return nil, errors.New("Book download info not found")
		}
	}
	return nil, errors.New("Not connected to DB")
}

func DbClose() error {
	if DatabaseObj.Connected {
		err := DatabaseObj.DbObj.Close()
		return err
	}
	return errors.New("Not connected")
}

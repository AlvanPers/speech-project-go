package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	/* чтобы не забыть. go к sql пути не видит без этого
	   go mod init knocker
	   go mod tidy*/
	_ "github.com/go-sql-driver/mysql"
)

/*структура идентичная базе данных*/
type Product struct {
	Id        int
	Headtext  string
	Bodytext  string
	Finaltext string
}

var database *sql.DB
var count int

/*выбор и вывод всех значений из базы данных*/
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	rows, err := database.Query("select * from text_expr.speech")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	speech := []Product{} /*speech - массив структур, в котором столбцы являются структурами Product*/

	for rows.Next() {
		p := Product{
			Id:        0,
			Headtext:  "",
			Bodytext:  "",
			Finaltext: "",
		}
		err := rows.Scan(&p.Id, &p.Headtext, &p.Bodytext, &p.Finaltext)
		if err != nil {
			fmt.Println(err)
			continue
		}
		speech = append(speech, p)
	}
	tmpl.Execute(w, speech)
}

/*Подсчёт строк в базе*/
func CountInBase() {
	err := database.QueryRow("SELECT COUNT(*) FROM text_expr.speech").Scan(&count)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		log.Println("Number of rows are =", count)
	}
}

/*----тут вывод рандомом полей базы-----*/
func IndexRandom(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/indexrandom.html")

	speech := []Product{}
	CountInBase()

	//вывод приветствия*******************
	rows, err := database.Query("SELECT Headtext FROM text_expr.speech WHERE Headtext != '' ORDER BY rand() LIMIT 1")

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		h := Product{
			//Id:       id,
			Headtext: "",
			//Bodytext:  "",
			//Finaltext: "",
		}
		err := rows.Scan( /*&h.Id,*/ &h.Headtext)
		if err != nil {
			log.Fatal(err)
			continue
		}
		speech = append(speech, h)
	}
	//*************************************************
	//вывод основного текста
	ArrNumber := make([]int, count)
	//заполнение массива от 1 до количества id(count) в базе
	for j := 0; j < count; j++ {
		ArrNumber[j] = j + 1
	}
	fmt.Println("massiv id = ", ArrNumber)

	//перемешиваем срез(массив)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(ArrNumber), func(i, j int) { ArrNumber[i], ArrNumber[j] = ArrNumber[j], ArrNumber[i] })
	//fmt.Printf("massiv id shuffle = %q\n", ArrNumber)
	fmt.Println("massiv id = ", len(ArrNumber))

	//проход цикла равен количеству строк(пока так, всё что есть) в базе
	for i := 0; i < len(ArrNumber); i++ {
		id := ArrNumber[i]
		rows, err := database.Query("select Id, Bodytext from text_expr.speech where id = ?", id)

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			p := Product{
				Id: id,
				//	Headtext:  "",
				Bodytext: "",
				//Finaltext: "",
			}
			err := rows.Scan(&p.Id, &p.Bodytext /*, &p.Finaltext*/)
			if err != nil {
				log.Fatal(err)
				continue
			}
			//log.Println(p.Id, p.Headtext)
			speech = append(speech, p) /*накапливаем массив(срез) структур product с помощью переменной p*/
		}
	}
	//вывод завершающей фразы****************
	rows, err = database.Query("SELECT Finaltext FROM text_expr.speech WHERE Finaltext != '' ORDER BY rand() LIMIT 1")

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		f := Product{
			//Id:       id,
			//Headtext: "",
			//Bodytext:  "",
			Finaltext: "",
		}
		err := rows.Scan( /*&h.Id,*/ &f.Finaltext)
		if err != nil {
			log.Fatal(err)
			continue
		}
		speech = append(speech, f)
	}
	//*********************************************
	tmpl.Execute(w, speech) /*вывод выбранных строк */

}

/*----------------*/
//добавление в базу фраз
func IndexAddInBase(w http.ResponseWriter, r *http.Request) {
	//	tmpl, _ := template.ParseFiles("templates/indexAddInBase.html")
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		Headtext := r.FormValue("Headtext")
		Bodytext := r.FormValue("Bodytext")
		Finaltext := r.FormValue("Finaltext")

		_, err = database.Exec("insert into text_expr.speech (Headtext, Bodytext, Finaltext) values (?, ?, ?)",
			Headtext, Bodytext, Finaltext)

		if err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	} else {
		http.ServeFile(w, r, "templates/indexAddInBase.html")
	}

}

/*реализация рандома*/
/*func randomInt(min, max int) int {return min + rand.Intn(max-min)}*/

/*объявление сервера и подключение к sql*/
func start() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3307)/text_expr")
	if err != nil {
		log.Println(err)
	}
	database = db
	defer db.Close()
	fmt.Println("Server is listening...")

	http.HandleFunc("/", IndexHandler)                       //начальная страница
	http.HandleFunc("/indexrandom.html/", IndexRandom)       //вывод рандомного текста
	http.HandleFunc("/indexAddInBase.html/", IndexAddInBase) //добавление в базу
	http.ListenAndServe(":8080", nil)
}

func main() {

	start()
}

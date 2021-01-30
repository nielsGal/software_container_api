package main

import (
	"github.com/gofiber/fiber"
	"github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"fmt"
	"time"
	"strconv"
	"os"
)

type Book struct{
	gorm.Model
	Title string `json:"title"`
	Price uint	`json:"price"`
	ISBN string	`json:"isbn"`
	Author string	`json:"author"`
}
type CartItem struct{
	gorm.Model
	Book Book `json:"book" gorm:"embedded"`  
	Quantity int `json:"quantity"`
	CartRefer int 
}
type Cart  struct {
	gorm.Model
	Token string `json:"token" gorm:"primaryKey"`
	Items []CartItem `json:"items" gorm:"ForeignKey:CartRefer"`
}
type IdRequest struct{
	Id int `json:"id"`
}
type CartRequest struct{
	Id int  `json:"id"`
	Token string `json:"token"`
	Quantity int  `json:"quantity"`
}

var db *gorm.DB

func main() {
	app := fiber.New()
	//setup the database
	setupDatabase()
	//use db conn to migrate
	db.AutoMigrate(&Book{})
	db.AutoMigrate(&CartItem{})
	db.AutoMigrate(&Cart{})
	//setup routes
	app.Get("/book/:id",getBook)
	app.Get("/books",getBooks)
	app.Post("/book",putBookInCart)
	app.Get("/token",getSessionToken)
	app.Get("/cart/:token",getCartItems)
	app.Post("/create",createBook)
	app.Listen(":3000")
}

func setupDatabase(){
	host := getEnv("POSTGRES_SERVICE_SERVICE_HOST","localhost")
	port := getEnv("POSTGRES_SERVICE_SERVICE_PORT","5432")
	user := getEnv("POSTGRES_USER","postgres")
	password := getEnv("POSTGRES_PASSWORD","")
	dbname := getEnv("POSGRES_DB","postgres")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Berlin",host,user,password,dbname,port)
	for true {
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if (err == nil) {
			break
		}
		
		fmt.Println("Connection failed. Retrying...")
		time.Sleep(time.Second)
	}
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func getBooks(fiberctx *fiber.Ctx) error {
	var books []Book
	booksfromDb := db.Find(&books)
	if booksfromDb.Error != nil{
		fmt.Println(booksfromDb.Error)
		return fiberctx.JSON(fiber.Map{
				"succes":false,
				"message":"could not find books",
			})
	}
	return fiberctx.JSON(books)
}

func getBook(fiberctx *fiber.Ctx) error{
	strid := fiberctx.Params("id")
	id ,err := strconv.Atoi(strid)
	if err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"could not find book with that id",
		})
	}
	book := new(Book)
	err = db.First(&book,id).Error
	if err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"could not find book with that id",
		})
	}
	return fiberctx.JSON(book)
}

func putBookInCart(fiberctx  *fiber.Ctx) error{
	cartRequest := new(CartRequest)
	//
	if err := fiberctx.BodyParser(cartRequest); err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"Something went wrong",
		})
	}
	//
	cart := new(Cart)
	err := db.Where("token = ?",cartRequest.Token).First(&cart).Error
	if err != nil {
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"could not find cart with that token",
		})
	}
	cartItem := new(CartItem)
	cartItem.Quantity = cartRequest.Quantity
	book := new(Book)

	err = db.First(&book,cartRequest.Id).Error
	if err != nil {
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"could not find  book with  that id",
		})
	}
	cartItem.Book = *book
	
	cart.Items = append(cart.Items,*cartItem)
	err = db.Save(&cart).Error
	if err != nil {
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"could not save to db",
		})
	}


	return fiberctx.JSON(fiber.Map{
		"success":true,
		"message":"added to cart",
	})
}

func getSessionToken(fiberctx *fiber.Ctx) error{
	uid := uuid.Must(uuid.NewV4())
	
	cart := new(Cart)
	cart.Token = uid.String()
	result := db.Create(&cart)
	if result.Error != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"error":result.Error,
		})
	}

	return fiberctx.JSON(fiber.Map{
		"success":true,
		"token":uid.String(),
	})
}

func getCartItems(fiberctx *fiber.Ctx) error{
	cartToken := fiberctx.Params("token")	
	cart := new(Cart)
	database := db.Where("token = ?",cartToken).First(&cart)
	err := database.Error
	fmt.Println(cart)
	fmt.Println(database.RowsAffected)
	if err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"error":"could not find cart",
		})
	}
	items := []CartItem{}
	db.Where("cart_refer = ?",cart.ID).Find(&items)
	print(items)
	return fiberctx.JSON(items)
}

func createBook(fiberctx *fiber.Ctx) error{
	book := new(Book)
	if err := fiberctx.BodyParser(book); err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"did not add the book",
		})
	}
	result := db.Create(book)
	if result.Error != nil {
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"message":"did not add the book",
		})
	}
	return fiberctx.JSON(fiber.Map{
		"success":true,
		"message":"successfully added book",
	})
}

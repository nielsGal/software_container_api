package main

import ("github.com/gofiber/fiber"
		"github.com/satori/go.uuid"
		"gorm.io/gorm"
		"context"
		"fmt"
		"gorm.io/driver/sqlite"
		"github.com/go-redis/redis"
		"strconv"
		"strings")

type Book struct{
	gorm.Model
	Title string `json:"title"`
	Price uint	`json:"price"`
	ISBN string	`json:"isbn"`
	Author string	`json:"author"`
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
var rdb *redis.Client
var ctx = context.Background()

func main() {
	app := fiber.New()
	var err error
	//setup the database
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("connection failed")
	}
	db.AutoMigrate(&Book{})
	//setup the redis
	rdb = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
	})
	//setup the routes
	app.Get("/book/:id",getBook)
	app.Get("/books",getBooks)
	app.Post("/book",putBookInCart)
	app.Get("/token",getSessionToken)
	app.Get("/cart/:token",getCartItems)
	app.Post("/create",createBook)
	app.Listen(":3000")
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
			"succes":false,
			"message":"Something went wrong",
		})
	}
	//
	cart, err := rdb.Get(ctx,cartRequest.Token).Result()
	cartMap := getMap(cart)
	if err!= nil{
		return fiberctx.JSON(fiber.Map{
			"succes":false,
			"message":"Something went wrong",
		})
	}
	//add something to the cart
	if val, ok := cartMap[cartRequest.Id]; ok{
		cartMap[cartRequest.Id] = val + cartRequest.Quantity
	}else{
		cartMap[cartRequest.Id] = cartRequest.Quantity
	}

	return fiberctx.JSON(fiberctx.JSON(fiber.Map{
		"succes":true,
		"message":"added to cart",
	}))
}

func getSessionToken(fiberctx *fiber.Ctx) error{
	uid := uuid.Must(uuid.NewV4())
	err := rdb.Set(ctx, uid.String(), "", 0).Err()
	if err != nil{
		return fiberctx.JSON(fiber.Map{
			"success":false,
			"token":"",
		})
	}
	return fiberctx.JSON(fiber.Map{
		"success":true,
		"token":uid.String(),
	})
}
func getCartItems(fiberctx *fiber.Ctx) error{
	cartToken := fiberctx.Params("token")

	return fiberctx.SendString(cartToken)
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

func getMap(s string) map[int]int {
	result := make(map[int]int)
	mapItems := strings.Split(s,"%")
	for i := 0 ; i <len(mapItems);i++{
		items := strings.Split(mapItems[i],":")
		quantity, err := strconv.Atoi(items[1])
		if err != nil{
			quantity = 0
		}
		id, err:= strconv.Atoi(items[0])
		if err == nil{
			result[id] = quantity
		}
	}
	return result
}
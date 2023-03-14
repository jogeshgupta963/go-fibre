package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db	*mongo.Database
}

var mg MongoInstance

const dbName = "fibre-hrms"
const mongoURI = "mongodb+srv://admin:FYv5jqnidHPCxCOr@testing.s5sej.mongodb.net/"+dbName+"?retryWrites=true&w=majority"
type Employee struct {
	ID     string   `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string	`json:"name"`
	Salary float64	`json:"salary"`
	Age    float64	`json:"age"`
}

func ConnectDB() error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println(err)
		return err
	}
	db := client.Database(dbName)
	mg = MongoInstance{
		Client: client,
		Db: db,
	}
	return nil
}

func getAllEmployee(c *fiber.Ctx) error {
	
	query:= bson.D{{}}

	cursor,err := mg.Db.Collection("employees").Find(c.Context(),query)
	if err!=nil{
		return c.Status(500).SendString(err.Error())
	}
	var emps []Employee 
	cursor.All(c.Context(),&emps)
	return c.JSON(emps)

}
func createEmployee(c *fiber.Ctx) error {
	collection := mg.Db.Collection("employees")

	emp := new(Employee)
	if err := c.BodyParser(emp); err != nil {
		return c.Status(400).SendString(err.Error())
	}	
	emp.ID = ""
	res,err := collection.InsertOne(c.Context(),emp)
	if err!=nil{
		return c.Status(500).SendString(err.Error())
	}
	filter := bson.D{{Key: "_id", Value: res.InsertedID}}
	createdRecord := collection.FindOne(c.Context(), filter)

	createdEmployee := &Employee{}
	createdRecord.Decode(createdEmployee)
	return c.Status(201).JSON(createdEmployee)
	
}

func updateEmployee(c *fiber.Ctx) error {
	collection :=mg.Db.Collection("employees")
	id := c.Params("id")
	empid,err := primitive.ObjectIDFromHex(id)
	if err!=nil{
		return c.Status(400).SendString(err.Error())
	}
	var emp Employee

	c.BodyParser(emp)
	
	query := bson.D{{Key: "_id",Value: empid}}
	update := bson.D{
		{
			Key:"$set",
			Value: bson.D{
				{Key: "name",Value:emp.Name},
				{Key: "age",Value:emp.Age},
				{Key: "salary",Value:emp.Salary},
			},
		},
	}

	err = collection.FindOneAndUpdate(c.Context(),query,update).Err()
	if err!=nil{
		if err == mongo.ErrNoDocuments{
			return c.SendStatus(400)
		}
	}
	emp.ID = id 
	return c.Status(200).JSON(emp);

}
func deleteEmployee(c *fiber.Ctx) error {

	employeeID, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		return c.SendStatus(400)
	}

	query := bson.D{{Key: "_id", Value: employeeID}}
	result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), &query)

	if err != nil {
		return c.SendStatus(500)
	}

	if result.DeletedCount < 1 {
		return c.SendStatus(404)
	}

	return c.Status(200).JSON("record deleted")

}
func main() {
	if err := ConnectDB(); err!=nil{
		log.Fatal("db error")
	}
	app := fiber.New()
	app.Get("/employee",getAllEmployee)
	app.Post("/employee",createEmployee)
	app.Put("/employee/:id",updateEmployee)
	app.Delete("/employee/:id",deleteEmployee)
	log.Fatal(app.Listen(":5000"))
}
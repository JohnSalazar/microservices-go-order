package repositories

import (
	"context"
	"encoding/json"
	"time"

	"order/src/models"

	"github.com/google/uuid"
	"github.com/oceano-dev/microservices-go-common/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository struct {
	database *mongo.Database
}

func NewOrderRepository(
	database *mongo.Database,
) *OrderRepository {
	return &OrderRepository{
		database: database,
	}
}

func (r *OrderRepository) collectionName() string {
	return "orders"
}

func (r *OrderRepository) collection() *mongo.Collection {
	return r.database.Collection(r.collectionName())
}

func (r *OrderRepository) find(ctx context.Context, filter interface{}) ([]*models.Order, error) {
	findOptions := options.FindOptions{}
	findOptions.SetSort(bson.M{"created_at": -1})

	newFilter := map[string]interface{}{
		"deleted": false,
	}
	mergeFilter := helpers.MergeFilters(newFilter, filter)

	cursor, err := r.collection().Find(ctx, mergeFilter, &findOptions)
	if err != nil {
		defer cursor.Close(ctx)
		return nil, err
	}

	orders := []*models.Order{}

	for cursor.Next(ctx) {
		object := map[string]interface{}{}

		err = cursor.Decode(object)
		if err != nil {
			return nil, err
		}

		order, err := r.mapOrder(object)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepository) findOne(ctx context.Context, filter interface{}) (*models.Order, error) {
	findOneOptions := options.FindOneOptions{}
	findOneOptions.SetSort(bson.M{"version": -1})

	newFilter := map[string]interface{}{
		"deleted": false,
	}
	mergeFilter := helpers.MergeFilters(newFilter, filter)

	object := map[string]interface{}{}
	err := r.collection().FindOne(ctx, mergeFilter, &findOneOptions).Decode(object)
	if err != nil {
		return nil, err
	}

	order, err := r.mapOrder(object)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) findOneAndUpdate(ctx context.Context, filter interface{}, fields interface{}) *mongo.SingleResult {
	findOneAndUpdateOptions := options.FindOneAndUpdateOptions{}
	findOneAndUpdateOptions.SetReturnDocument(options.After)

	result := r.collection().FindOneAndUpdate(ctx, filter, bson.M{"$set": fields}, &findOneAndUpdateOptions)

	return result
}

func (r *OrderRepository) GetAll(ctx context.Context, customerID primitive.ObjectID) ([]*models.Order, error) {
	filter := bson.M{"customer_id": customerID}

	return r.find(ctx, filter)
}

func (r *OrderRepository) FindByCustomerID(ctx context.Context, customerID primitive.ObjectID) (*models.Order, error) {
	filter := bson.M{"customer_id": customerID}

	return r.findOne(ctx, filter)
}

func (r *OrderRepository) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.Order, error) {
	filter := bson.M{"_id": ID}

	return r.findOne(ctx, filter)
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	products := r.mapOrderProducts(order.Products)

	fields := bson.M{
		"_id":         order.ID,
		"customer_id": order.CustomerID,
		"products":    products,
		"sum":         order.Sum,
		"discount":    order.Discount,
		"status":      order.Status,
		"status_at":   order.StatusAt,
		"created_at":  time.Now().UTC(),
		"version":     0,
		"deleted":     false,
	}

	_, err := r.collection().InsertOne(ctx, fields)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *models.Order) (*models.Order, error) {
	order.Version++
	order.UpdatedAt = time.Now().UTC()

	products := r.mapOrderProducts(order.Products)
	stores := r.mapOrderStores(order.Stores)

	fields := bson.M{
		"products":   products,
		"stores":     stores,
		"sum":        order.Sum,
		"discount":   order.Discount,
		"status":     order.Status,
		"status_at":  order.StatusAt,
		"updated_at": order.UpdatedAt,
		"version":    order.Version,
	}

	filter := r.filterUpdate(order)

	result := r.findOneAndUpdate(ctx, filter, fields)
	if result.Err() != nil {
		return nil, result.Err()
	}

	object := map[string]interface{}{}
	err := result.Decode(object)
	if err != nil {
		return nil, err
	}

	modelOrder, err := r.mapOrder(object)
	if err != nil {
		return nil, err
	}

	return modelOrder, err
}

func (r *OrderRepository) Delete(ctx context.Context, ID primitive.ObjectID) error {
	filter := bson.M{"_id": ID}

	fields := bson.M{"deleted": true}

	result := r.findOneAndUpdate(ctx, filter, fields)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func (r *OrderRepository) filterUpdate(order *models.Order) interface{} {
	filter := bson.M{
		"_id":     order.ID,
		"version": order.Version - 1,
	}

	return filter
}

func (r *OrderRepository) mapOrder(object map[string]interface{}) (*models.Order, error) {
	jsonStr, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var order models.Order
	if err := json.Unmarshal(jsonStr, &order); err != nil {
		return nil, err
	}

	order.ID = object["_id"].(primitive.ObjectID)
	order.CustomerID = object["customer_id"].(primitive.ObjectID)

	if object["products"] != nil {
		var products []*models.Product
		listProducts := object["products"].(primitive.A)
		for _, product := range listProducts {
			product, err := r.mapProductFromInterfaceToModel(product.(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			products = append(products, product)
		}
		order.Products = products
	}

	if object["stores"] != nil {
		var stores []*models.Store
		listStores := object["stores"].(primitive.A)
		for _, store := range listStores {
			store, err := r.mapStoreFromInterfaceToModel(store.(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			stores = append(stores, store)
		}
		order.Stores = stores
	}

	return &order, nil
}

func (r *OrderRepository) mapProductFromInterfaceToModel(object map[string]interface{}) (*models.Product, error) {
	jsonStr, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal(jsonStr, &product); err != nil {
		return nil, err
	}

	ID, err := uuid.Parse(object["_id"].(string))
	if err != nil {
		return nil, err
	}
	product.ID = ID

	return &product, nil
}

func (r *OrderRepository) mapStoreFromInterfaceToModel(object map[string]interface{}) (*models.Store, error) {
	store := models.Store{}

	ID, err := uuid.Parse(object["_id"].(string))
	if err != nil {
		return nil, err
	}

	ProductID, err := uuid.Parse(object["product_id"].(string))
	if err != nil {
		return nil, err
	}

	store.ID = ID
	store.ProductID = ProductID

	return &store, nil
}

func (r *OrderRepository) mapOrderProducts(orderProducts []*models.Product) []map[string]interface{} {
	var products []map[string]interface{}
	for _, product := range orderProducts {
		modelProduct := map[string]interface{}{
			"_id":         product.ID.String(),
			"Name":        product.Name,
			"Description": product.Description,
			"Price":       product.Price,
			"Quantity":    product.Quantity,
			"Image":       product.Image,
		}

		products = append(products, modelProduct)
	}

	return products
}

func (r *OrderRepository) mapOrderStores(orderStores []*models.Store) []map[string]interface{} {
	var stores []map[string]interface{}
	for _, store := range orderStores {
		modelStore := map[string]interface{}{
			"_id":        store.ID.String(),
			"product_id": store.ProductID.String(),
		}

		stores = append(stores, modelStore)
	}

	return stores
}

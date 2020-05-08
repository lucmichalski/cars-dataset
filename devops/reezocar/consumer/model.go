package main

type Announce struct {
	AnnounceId				string	`bson:"announce_id" json:"announce_id"`
	Title					string	`bson:"title" json:"title"`
	URL						string	`bson:"url" json:"url"`
	Price					int		`bson:"prince" json:"price"`
	Phone					string	`bson:"phone" json:"phone"`
	GarageName				string	`bson:"garage_name" json:"garage_name"`
	GarageAddress			string	`bson:"garage_address" json:"garage_address"`
	CarMileage				string	`bson:"car_mileage" json:"car_mileage"`
	CarFirstRegistration	string	`bson:"car_first_registration" json:"car_first_registration"`
	AnnouncePostalCode		string	`bson:"announce_postal_code" json:"announce_postal_code"`
	CarEngine				string	`bson:"car_engine" json:"car_engine"`
	Transmission			string	`bson:"transmission" json:"transmission"`
	FiscalPower				string	`bson:"fiscal_power" json:"fiscal_power"`
}
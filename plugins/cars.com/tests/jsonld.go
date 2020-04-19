package main

import (
	"encoding/json"
	"log"
	"fmt"
	"sort"

	"github.com/k0kubun/pp"	
	"github.com/astaxie/flatmap"
)

func main() {

	var carInfo map[string]interface{}
	docStr := `{
  "jsonld": [
    {
      "@context": "http://schema.org",
      "@type": "Car",
      "driveWheelConfiguration": {
        "@type": "DriveWheelConfigurationValue",
        "name": "AWD"
      },
      "fuelEfficiency": {
        "@type": "QuantitativeValue",
        "unitCode": "MPG",
        "value": ""
      },
      "fuelType": "E85 Flex Fuel",
      "mileageFromOdometer": {
        "@type": "QuantitativeValue",
        "unitCode": "SMI",
        "value": 65448
      },
      "numberOfDoors": 4,
      "vehicleEngine": {
        "@type": "EngineSpecification",
        "name": "3.6L V6 24V MPFI DOHC Flexible Fuel"
      },
      "vehicleIdentificationNumber": "1C3CCCEGXFN540869",
      "vehicleInteriorColor": "",
      "vehicleModelDate": 2015,
      "vehicleSeatingCapacity": "5",
      "vehicleTransmission": "9-Speed Automatic",
      "color": "Black",
      "itemCondition": {
        "@type": "OfferItemCondition",
        "name": "Used"
      },
      "manufacturer": {
        "@type": "Organization",
        "name": "Chrysler"
      },
      "model": {
        "@type": "ProductModel",
        "name": "2015 Chrysler 200 C",
        "isVariantOf": "https://www.cars.com/research/chrysler-200-2015/"
      },
      "offers": {
        "@type": "Offer",
        "price": 14999,
        "priceCurrency": "USD",
        "seller": {
          "@type": "Organization",
          "name": "Main Motorcar Chrysler Dodge Jeep RAM",
          "telephone": "(518) 212-3426",
          "address": {
            "addressLocality": "Johnstown",
            "addressRegion": "NY",
            "streetAddress": "224-228 W Main St"
          },
          "aggregateRating": {
            "@type": "AggregateRating",
            "ratingValue": 4.3,
            "reviewCount": 19
          }
        }
      },
      "image": {
        "@type": "ImageObject",
        "contentUrl": "https://www.cstatic-images.com/phototab/in/v1/455036/1C3CCCEGXFN540869/31e34303dd3ab62c0d095564c9227061.jpg"
      },
      "name": "2015 Chrysler 200 C"
    },
    {
      "@context": "http://schema.org",
      "@type": "BreadcrumbList",
      "itemListElement": [
        {
          "@type": "ListItem",
          "position": 1,
          "item": {
            "@id": "https://www.cars.com/",
            "name": "Cars.com"
          }
        },
        {
          "@type": "ListItem",
          "position": 2,
          "item": {
            "@id": "https://www.cars.com/shopping/",
            "name": "Shop"
          }
        },
        {
          "@type": "ListItem",
          "position": 3,
          "item": {
            "@id": "https://www.cars.com/shopping/chrysler-200-2015/",
            "name": "2015 Chrysler 200"
          }
        },
        {
          "@type": "ListItem",
          "position": 4,
          "item": {
            "@id": "https://www.cars.com/vehicledetail/detail/804122829/overview/",
            "name": "VIN: 1C3CCCEGXFN540869"
          }
        }
      ]
    },
    {
      "@context": "http://schema.org",
      "@type": "WebSite",
      "name": "Cars.com",
      "url": "https://www.cars.com/"
    },
    {
      "@context": "http://schema.org",
      "@type": "Organization",
      "url": "https://www.cars.com/",
      "logo": "https://graphics.cars.com/images/core/logo.png",
      "sameAs": [
        "https://www.facebook.com/CarsDotCom/",
        "https://www.twitter.com/carsdotcom",
        "https://www.pinterest.com/carsdotcom/",
        "https://www.youtube.com/user/Carscom",
        "https://instagram.com/carsdotcom/",
        "https://www.linkedin.com/company/cars-com",
        "https://plus.google.com/+CarsDotCom",
        "https://en.wikipedia.org/wiki/Cars.com"
      ]
    }
  ]
}`

	if err := json.Unmarshal([]byte(docStr), &carInfo); err != nil {
		log.Fatalln("unmarshal error, ", err)
	}

	pp.Println(carInfo)

	fm, err := flatmap.Flatten(carInfo)
	if err != nil {
		log.Fatal(err)
	}
	var ks []string
	for k :=range fm {
		ks = append(ks,k)		
	}
	sort.Strings(ks)
	for _, k :=range ks {
		fmt.Println(k,":",fm[k])
	}

	pp.Println(fm["jsonld.0.model.name"])
	pp.Println(fm["jsonld.0.@context"])

}

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/imc-trading/dock2box/d2bsrv/models"
	"github.com/imc-trading/dock2box/d2bsrv/version"
)

type SubnetController struct {
	database string
	session  *mgo.Session
}

func NewSubnetController(s *mgo.Session) *SubnetController {
	return &SubnetController{
		database: "d2b",
		session:  s,
	}
}

func (c SubnetController) SetDatabase(database string) {
	c.database = database
}

func (c SubnetController) CreateIndex() {
	index := mgo.Index{
		Key:    []string{"subnet"},
		Unique: true,
	}

	if err := c.session.DB(c.database).C("subnets").EnsureIndex(index); err != nil {
		panic(err)
	}
}

func (c SubnetController) All(w http.ResponseWriter, r *http.Request) {
	// Initialize empty struct list
	s := []models.Subnet{}

	// Get all entries
	if err := c.session.DB(c.database).C("subnets").Find(nil).All(&s); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.URL.Query().Get("embed") == "true" {
		for i, v := range s {
			// Get site
			if err := c.session.DB(c.database).C("sites").FindId(v.SiteID).One(&s[i].Site); err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}

	// Write content-type, header and payload
	jsonWriter(w, r, s, http.StatusOK)
}

func (c SubnetController) Get(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	prefix := mux.Vars(r)["prefix"]

	// Initialize empty struct
	s := models.Subnet{}

	// Get entry
	if err := c.session.DB(c.database).C("subnets").Find(bson.M{"subnet": name + "/" + prefix}).One(&s); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.URL.Query().Get("embed") == "true" {
		// Get site
		if err := c.session.DB(c.database).C("sites").FindId(s.SiteID).One(&s.Site); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Write content-type, header and payload
	jsonWriter(w, r, s, http.StatusOK)
}

func (c SubnetController) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Validate ObjectId
	if !bson.IsObjectIdHex(id) {
		//		jsonError(w, r, fmt.Errorf("Incorrectly formated ID: %s", id), http.StatusInternalServerError)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get ID
	oid := bson.ObjectIdHex(id)

	// Initialize empty struct
	s := models.Subnet{}

	// Get entry
	if err := c.session.DB(c.database).C("subnets").FindId(oid).One(&s); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.URL.Query().Get("embed") == "true" {
		// Get site
		if err := c.session.DB(c.database).C("sites").FindId(s.SiteID).One(&s.Site); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Write content-type, header and payload
	jsonWriter(w, r, s, http.StatusOK)
}

func (c SubnetController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize empty struct
	s := models.Subnet{}

	// Decode JSON into struct
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		jsonError(w, r, "Failed to deconde JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set ID
	s.ID = bson.NewObjectId()

	// Validate input using JSON Schema
	docLoader := gojsonschema.NewGoLoader(s)
	schemaLoader := gojsonschema.NewReferenceLoader("http://localhost:8080/" + version.APIVersion + "/schemas/subnet.json")

	res, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		jsonError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if !res.Valid() {
		var errors []string
		for _, e := range res.Errors() {
			errors = append(errors, fmt.Sprintf("%s: %s", e.Context().String(), e.Description()))
		}
		jsonError(w, r, errors, http.StatusInternalServerError)
		return
	}

	// Set refs
	s.SiteRef = "/subnets/id/" + s.SiteID.Hex()

	// Insert entry
	if err := c.session.DB(c.database).C("subnets").Insert(s); err != nil {
		jsonError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write content-type, header and payload
	jsonWriter(w, r, s, http.StatusCreated)
}

func (c SubnetController) Remove(w http.ResponseWriter, r *http.Request) {
	// Get name
	name := mux.Vars(r)["name"]
	prefix := mux.Vars(r)["prefix"]

	// Remove entry
	if err := c.session.DB(c.database).C("subnets").Remove(bson.M{"subnet": name + "/" + prefix}); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Write status
	jsonWriter(w, r, nil, http.StatusOK)
}

func (c SubnetController) RemoveByID(w http.ResponseWriter, r *http.Request) {
	// Get ID
	id := mux.Vars(r)["id"]

	// Validate ObjectId
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get new ID
	oid := bson.ObjectIdHex(id)

	// Remove entry
	if err := c.session.DB(c.database).C("subnets").RemoveId(oid); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Write status
	jsonWriter(w, r, nil, http.StatusOK)
}

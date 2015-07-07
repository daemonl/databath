package xml_model

import (
	"fmt"

	"github.com/daemonl/databath"
	"github.com/daemonl/databath/types"
)

type Model struct {
	Collections      []*Collection `xml:"collection"`
	CustomQueries    []*Query      `xml:"query"`
	DynamicFunctions []*Function   `xml:"function"`
	Hooks            []*Hook       `xml:"hook"`
}

type Function struct {
	Name     string `xml:"name,attr"`
	Filename string `xml:"filename,attr"`
}

type View struct{}
type CustomField struct{}
type CollectionMask struct{}

type Query struct {
	Name      string           `xml:"name,attr"`
	Type      string           `xml:"type,attr"`
	SQL       string           `xml:"sql"`
	InFields  []QueryParameter `xml:"parameter"`
	OutFields []QueryField     `xml:"column"`
}

type QueryParameter struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type QueryField struct {
	Name string              `xml:"name,attr"`
	Type types.FieldTypeName `xml:"type,attr"`
}

type SearchPrefix struct {
	Prefix string `xml:"prefix,attr"`
	Field  string `xml:"field,attr"`
}
type Hook struct{}

func (m *Model) ToDatabath() (*databath.Model, error) {
	dbm := &databath.Model{}

	dbm.Collections = map[string]*databath.Collection{}
	for _, rawCollection := range m.Collections {
		collection, err := rawCollection.ToDatabath()
		if err != nil {
			return nil, err
		}
		collection.Model = dbm
		dbm.Collections[rawCollection.Name] = collection
	}

	// Should be in model.init or something

	for _, collection := range dbm.Collections {
		for path, field := range collection.Fields {
			field.Collection = collection
			field.Path = path

			refField, isRefField := field.FieldType.(*types.FieldRef)
			if !isRefField {
				continue
			}
			_, ok := dbm.Collections[refField.Collection]
			if !ok {
				return nil, fmt.Errorf("ref field %s.%s references collection %s, which doesn't exist", collection.TableName, path, refField.Collection)
			}
			//fmt.Printf("Foreign Key %s.%s -> %s\n", collection.TableName, field.Path, refField.Collection)
			dbm.Collections[refField.Collection].ForeignKeys = append(dbm.Collections[refField.Collection].ForeignKeys, field)
		}

	}

	return dbm, nil
}

/*
	collections := make(map[string]*Collection)

	for collectionName, rawCollection := range model.Collections {

		customFields := make(map[string]FieldSetFieldDef)

		for name, rawCustomField := range rawCollection.CustomFields {
			fsfd, err := getFieldSetFieldDef(name, rawCustomField)
			if err != nil {
				err = fmt.Errorf("in collection %s: %s", collectionName, err.Error())
				log.Printf(err.Error())
				return nil, err
			}
			customFields[name] = fsfd
		}

		fieldSets := make(map[string][]FieldSetFieldDef)
		if doFieldSets {

			if rawCollection.FieldSets == nil {
				rawCollection.FieldSets = make(map[string][]string)
			}

			_, hasDefaultFieldset := rawCollection.FieldSets["default"]
			if !hasDefaultFieldset {
				allFieldNames := make([]string, 0, 0)
				for fieldName, _ := range rawCollection.Fields {
					allFieldNames = append(allFieldNames, fieldName)
				}
				rawCollection.FieldSets["default"] = allFieldNames

			}

			_, hasIdentityFieldset := rawCollection.FieldSets["identity"]
			if !hasIdentityFieldset {
				_, exists := rawCollection.Fields["name"]
				if !exists {
					return nil, (fmt.Errorf("%s: No identity fieldset or 'name' field to fall back on.", collectionName))
				}

				rawCollection.FieldSets["identity"] = []string{"name"}
			}

			for name, rawSet := range rawCollection.FieldSets {
				rawSet = append(rawSet, "id")

				fieldSetDefs := make([]FieldSetFieldDef, len(rawSet), len(rawSet))
				for i, fieldName := range rawSet {
					if fieldName[0:1] == "-" {
						fieldName = fieldName[1:]
					}

					fieldName = strings.Split(fieldName, " ")[0]

					customField, ok := customFields[fieldName]
					if ok {
						fieldSetDefs[i] = customField
						continue
					}

					fsfd := FieldSetFieldDefNormal{
						path:      fieldName,
						pathSplit: strings.Split(fieldName, "."),
					}
					fieldSetDefs[i] = &fsfd

					//return nil, UserErrorF("No field or custom field for %s in %s", fieldName, collectionName)

				}
				fieldSets[name] = fieldSetDefs
			}
		}

		searchPrefixes := make(map[string]*SearchPrefix)
		for prefixStr, rawPrefix := range rawCollection.SearchPrefixes {
			//field, ok := fields[rawPrefix.Field]
			//if !ok {
			//	return nil, ParseErrF("Prefix referenced field '%s' which doesn't exist", rawPrefix.Field)
			//}
			prefix := SearchPrefix{
				//Field:     field,
				Prefix:    prefixStr,
				FieldName: rawPrefix.Field,
			}
			searchPrefixes[prefixStr] = &prefix
		}

		masks := map[uint64]*Mask{}

		for users, rawMask := range rawCollection.Masks {

			r, rok := rawMask["read"]
			w, wok := rawMask["write"]

			mask := &Mask{}
			if rok {
				mask.Read = make([]string, len(r), len(r))
				for i, name := range r {
					str, ok := name.(string)
					if !ok {
						return nil, ParseErrF("Mask fieldset name not string")
					}
					mask.Read[i] = str
				}
			}
			if wok {
				mask.Write = make([]string, len(r), len(r))
				for i, name := range w {
					str, ok := name.(string)
					if !ok {
						return nil, ParseErrF("Mask fieldset name not string")
					}
					mask.Write[i] = str
				}
			}

			for _, uPart := range strings.Split(users, ",") {
				subUParts := strings.Split(uPart, "-")
				switch len(subUParts) {
				case 1:
					asInt, err := strconv.ParseUint(subUParts[0], 10, 64)
					if err != nil {
						return nil, ParseErrF("Mask identifier invalid %s", uPart)
					}
					masks[asInt] = mask

				case 2:
					asIntFrom, err1 := strconv.ParseUint(subUParts[0], 10, 64)
					asIntTo, err2 := strconv.ParseUint(subUParts[0], 10, 64)
					if err1 != nil || err2 != nil || asIntFrom > asIntTo {
						return nil, ParseErrF("Mask identifier invalid %s", uPart)
					}

					for i := asIntFrom; i <= asIntTo; i++ {
						masks[i] = mask
					}

				}
			}
		}

		collection := Collection{
			Fields:         fields,
			FieldSets:      fieldSets,
			TableName:      collectionName,
			ForeignKeys:    make([]*Field, 0, 0),
			CustomFields:   customFields,
			SearchPrefixes: searchPrefixes,
			Masks:          masks,
			ViewQuery:      rawCollection.ViewQuery,
		}

		collections[collectionName] = &collection
	}

	dynamicFunctions := model.DynamicFunctions

	customQueries := make(map[string]*CustomQuery)
	for queryName, rawQuery := range model.CustomQueries {
		cq := CustomQuery{
			Query:     rawQuery.Query,
			InFields:  make([]*Field, len(rawQuery.InFields), len(rawQuery.InFields)),
			OutFields: make(map[string]*Field),
			Type:      rawQuery.Type,
		}
		for i, rawField := range rawQuery.InFields {
			field, err := FieldFromDef(rawField)
			if err != nil {
				return nil, (fmt.Errorf("Error parsing Raw Query %s.[in][%d] - %s", queryName, i, err.Error()))
			}
			cq.InFields[i] = field
		}
		for i, rawField := range rawQuery.OutFields {
			field, err := FieldFromDef(rawField)
			if err != nil {
				return nil, (fmt.Errorf("Error parsing Raw Query %s.[out][%d] - %s", queryName, i, err.Error()))
			}
			cq.OutFields[i] = field
		}
		customQueries[queryName] = &cq
	}
	for _, h := range model.Hooks {

		if h.Raw != nil {
			rawQuery := h.Raw
			cq := CustomQuery{
				Query:     rawQuery.Query,
				InFields:  make([]*Field, len(rawQuery.InFields), len(rawQuery.InFields)),
				OutFields: make(map[string]*Field),
				Type:      rawQuery.Type,
			}
			for i, rawField := range rawQuery.InFields {
				field, err := FieldFromDef(rawField)
				if err != nil {
					log.Println(err)
					return nil, (fmt.Errorf("Error parsing hook ", err.Error()))
				}
				cq.InFields[i] = field
			}
			h.CustomAction = &cq
		}

		collection, ok := collections[h.Collection]
		if !ok {
			return nil, UserErrorF("Hook on non existing collection %s", h.Collection)
		}
		collection.Hooks = append(collection.Hooks, h)

	}

	returnModel := &Model{
		Collections:      collections,
		CustomQueries:    customQueries,
		DynamicFunctions: dynamicFunctions,
	}


	return returnModel, err
}

}*/

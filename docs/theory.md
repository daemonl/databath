


SELECT person.name, organisation.phone
FROM person
LEFT JOIN organisation ON organisation.id = person.organisation
WHERE ...


root = query(personCollection)
organisation = root.LeftJoin(organisation)

root.Add(name)
organisation.Add(phone)

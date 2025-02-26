# TODO
* Write the README with a feature list section going over the files
## Features
* Decouple the revocation middleware querying the database in the blacklist service
* Adding CORS(chi/cors) middleware to use the API and frontend on different (sub)domains
* CRUD operations transactions with rollback
* Unit testing
* Handlers context timeout
* Possible often used SQL tables indexing
## Adding more SQL databases support
* MySQL and MariaDB
* Oracle
* Microsoft
* CockroachDB
* Clickhouse
* Amazon Redshift
## Auditing even more
* BOLA(Broken Object Level Authorization) and IDOR(Insecure Direct Object Reference)
* Injections(shouldn't be happening because of paramatirazion)
* DOS(Denial Of Service) run a stress test on a production instance
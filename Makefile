up:
	docker-compose up
psql:
	docker-compose exec db psql --host=localhost --port=5432 --dbname=wg_forge_db --username=wg_forge --password

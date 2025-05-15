up:
    docker build -t pdfgen:latest .
    docker rm -f $(docker ps -aq --filter "label=app=pdfgen") || true
    docker stack deploy -c docker-compose.yml pdfgen

term:
    @docker exec -it $(docker ps --filter "label=app=pdfgen" -q) bash

front:
    @ CONTAINER_COUNT=$(docker ps | grep pdfgen | wc -l); \
    if [ "$CONTAINER_COUNT" -eq 0 ]; then \
        docker stack deploy -c docker-compose.yml pdfgen; \
    fi; \
    open "http://$(ipconfig getifaddr en0):8081"

logs:
    docker logs $(docker ps --filter "label=app=pdfgen" -q) -f

reload:
    just up
    just front
    just logs

test:
    docker compose -f docker-compose-test.yml up --build
    docker compose -f docker-compose-test.yml down

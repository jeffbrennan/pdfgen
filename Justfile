up:
    docker compose up --build -d

term:
    docker run -it --entrypoint /bin/bash pdfgen

front:
    @ CONTAINER_COUNT=$(docker ps | grep pdfgen | wc -l); \
    if [ "$CONTAINER_COUNT" -eq 0 ]; then \
        docker run -d --name pdfgen -p 8081:8081 pdfgen; \
    fi; \
    open "http://$(ipconfig getifaddr en0):8081"

logs:
    docker logs -f pdfgen

reload:
    just up
    just front
    just logs

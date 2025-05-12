build:
    docker build -t pdfgen .

run:
    docker run -it pdfgen

term:
    docker run -it --entrypoint /bin/bash pdfgen

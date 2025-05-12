FROM python:3.12-slim-bookworm

RUN apt-get update && apt-get install -y \
    wget \
    tar \
    vim \
    git \
    # latex
    texlive-latex-base \
    texlive-fonts-recommended \
    texlive-fonts-extra \
    texlive-latex-extra \
    # sphinx dependencies
    gcc \
    libkrb5-dev \
    libenchant-2-dev && \
    rm -rf /var/lib/apt/lists/*

RUN wget https://go.dev/dl/go1.22.12.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.22.12.linux-amd64.tar.gz && \
    rm go1.22.12.linux-amd64.tar.gz

# Add Go to PATH
ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

COPY pyproject.toml uv.lock* ./
RUN uv sync --locked

WORKDIR /build/src/pdfgen
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o .

EXPOSE 8081

CMD ["./pdfgen"]
# Pinger - Simple API for Site & Server Status

**Pinger** is a small but useful utility (microservice) that allows you to check if a website or server is running just by sending a request in your browser.

It can:
1. **Ping a server** (check response latency).
2. **Check HTTP status** (e.g., does the site work via http://).
3. **Check HTTPS status** (for secure connections).

In return, you get a convenient JSON response that is easy to use in your scripts, bots, or monitoring systems.

---

## 🚀 How to Use (API)

The service works like a website. You send it parameters, and it answers you.

### Request Parameters
- `host` (required): The website address or server IP you want to check.
- `method` (optional): The check method.
  - `ping` (default) — Standard ping.
  - `http` — Check http:// address.
  - `https` — Check https:// address.
- `key` (optional): Secret key, if set during launch (to protect against unauthorized access).

### Examples

**1. Ping a server (google.com)**
Request:
```
http://localhost:8088/?host=google.com
```
Response (JSON):
```json
{
  "host": "google.com",
  "type": "ping",
  "result": 14.2  // Average response time in milliseconds
}
```
*(If the server is unreachable, result will be 0)*

**2. Check site response code (HTTP status)**
Request:
```
http://localhost:8088/?host=google.com&method=https
```
Response:
```json
{
  "host": "google.com",
  "type": "https",
  "result": 200  // Code 200 means "OK"
}
```

---

## 🐳 How to Run with Docker (Easiest Way)

If you are completely new to this, think of Docker as a "virtual box" where the program runs perfectly without breaking anything on your computer.

You just need to install Docker Desktop and run one command in your terminal (PowerShell or CMD).

### 1. Standard Run
This command downloads our program (image `fedorananin/pinger`) and starts it.

```bash
docker run -d -p 8088:80 --name pinger --restart unless-stopped fedorananin/pinger
```

**What these letters mean:**
- `docker run`: "Hey Docker, run the program!"
- `-d`: "Run in background" (so it doesn't clutter your terminal).
- `-p 8088:80`: "Redirect port". Inside the box, the program listens on port 80, but we will access it via port 8088 on our computer.
- `--name my-pinger`: Give it the name "my-pinger" to make it easier to find later.
- `fedorananin/pinger`: The name of the image to download and run.

Now open in your browser: `http://localhost:8088/?host=google.com` — and it works!

### 2. Run with Password (Protection)
If you don't want just anyone using your pinger, add a secret key.

```bash
docker run -d -p 8088:80 --name pinger -e API_KEY=supersecret123 --restart unless-stopped fedorananin/pinger
```

- `-e API_KEY=supersecret123`: Creates an environment variable with your password.

Now requests without the key will not work. You need to add `&key=supersecret123`:
`http://localhost:8088/?host=google.com&key=supersecret123`

### Environment Variables

- `API_KEY` (optional): If set, all requests must include a matching `key` query parameter for authentication.
- `CONCURRENCY_LIMIT` (optional): Limits the number of concurrent ping/HTTP checks. Defaults to `20`. Set a lower value if your server has limited resources, or a higher value if you have plenty and expect high load.

For example, to run with an API key and a concurrency limit of 10:
```bash
docker run -d -p 8088:80 --name pinger -e API_KEY=supersecret123 -e CONCURRENCY_LIMIT=10 --restart unless-stopped fedorananin/pinger
```

---

## 🛠 How to Build and Run Yourself (Without Docker)

If you are a programmer or want to run this directly on your computer without "boxes".

### Requirements
1. Installed **Go** (Golang) version 1.21 or higher.
2. Installed `ping` utility in the system (usually available on Linux/Mac, and should be in paths on Windows).

### Instructions
1. Download the source code.
2. Open a terminal in the code folder.
3. Run:
   ```bash
   go run main.go
   ```
   Or build an `.exe` file (or binary for Linux):
   ```bash
   go build -o pinger main.go
   ./pinger
   ```

The server will start on port **80** (note: on Linux this often requires root/sudo rights, or change the port in the code).
If you want to set a protection key locally:
- **Windows (PowerShell):** `$env:API_KEY="mykey"; go run main.go`
- **Linux/Mac:** `export API_KEY=mykey && go run main.go`

---

## 📦 Building Your Own Docker Image

If you want to "package" changes into Docker yourself:

```bash
docker build -t my-custom-pinger .
```
Where `.` means "current folder". After that, you can run your image `my-custom-pinger` instead of `fedorananin/pinger`.

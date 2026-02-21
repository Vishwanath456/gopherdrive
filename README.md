<h1 align="center">🚀 GopherDrive</h1>
<p align="center">
High-Concurrency Metadata Processing System built with Go, gRPC & MySQL
</p>

<hr>

<h2>📊 Dashboard Preview</h2>
<p>
The dashboard provides real-time visibility into total files, completed jobs, processing status, and system health.
</p>

<p align="center">
<img width="402" height="162" alt="image" src="https://github.com/user-attachments/assets/e4883254-6177-49b7-8322-8b1fce6ce54c" />
</p>
<p align="center">
<img width="402" height="127" alt="image" src="https://github.com/user-attachments/assets/4525e844-13d4-4167-a4ff-93226d8d23df" />
</p>

<hr>

<h2>📤 Upload File (POST /files)</h2>
<p>
Files are uploaded via REST API, streamed to disk using <b>io.Copy</b>, renamed using <b>UUID</b>, and registered via gRPC.
</p>

<p align="center">
<img width="452" height="319" alt="image" src="https://github.com/user-attachments/assets/69bfbb0b-7ec6-451b-941f-af264602c9f4" />
</p>

<hr>

<h2>📥 Get File Metadata (GET /files/{id})</h2>
<p>
Retrieves file metadata including SHA256 hash, size, extension, and status through the gRPC service layer.
</p>

<p align="center">
<img width="358" height="260" alt="image" src="https://github.com/user-attachments/assets/e3d1e217-49a8-4d21-9553-b463ac17156f" />
</p>

<hr>

<h2>🩺 Health Check</h2>
<p>
The <code>/healthz</code> endpoint verifies database and gRPC connectivity to ensure production readiness.
</p>

<p align="center">
<img width="387" height="94" alt="image" src="https://github.com/user-attachments/assets/3296acb1-1ea2-480a-8b84-0329e8a7ea4b" />
</p>

<hr>

<h2>🛑 Graceful Shutdown</h2>
<p>
Handles SIGINT/SIGTERM signals and allows active worker jobs to complete before shutting down the server.
</p>

<p align="center">
<img width="452" height="23" alt="image" src="https://github.com/user-attachments/assets/eb619379-2d5d-4244-aca2-75cb3932cc25" />
</p>

<hr>

<h2>⚙️ Key Features</h2>

<ul>
  <li>✅ Bounded Worker Pool (5 Goroutines)</li>
  <li>✅ Asynchronous SHA256 Processing</li>
  <li>✅ gRPC-Based Metadata Service</li>
  <li>✅ MySQL Persistence Layer</li>
  <li>✅ REST Gateway</li>
  <li>✅ UUID-based Unique File Naming</li>
  <li>✅ Structured Logging (slog)</li>
  <li>✅ Health Check Endpoint</li>
  <li>✅ Graceful Shutdown</li>
</ul>

<hr>

<h2>▶️ How To Run</h2>

<pre>
go mod tidy
go run cmd/server/main.go
</pre>

<p>
Open in browser:
</p>

<pre>
http://localhost:8080
</pre>

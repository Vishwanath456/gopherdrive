async function loadFiles() {
    console.log("Dashboard JS loaded");
    try {
        const response = await fetch("/api/files");

        if (!response.ok) {
            console.error("API error:", response.status);
            return;
        }

        const files = await response.json();
        console.log("Files received:", files);

        const table = document.getElementById("fileTable");
        table.innerHTML = "";

        let completed = 0;
        let processing = 0;

        files.forEach(file => {

            // Support both uppercase and lowercase JSON
            const id = file.ID || file.id || "-";
            const status = file.Status || file.status || "-";
            const size = file.Size || file.size || "-";
            const extension = file.Extension || file.extension || "-";
            const sha = file.SHA256 || file.sha256 || "-";

            if (status === "completed") completed++;
            if (status === "processing") processing++;

            const row = document.createElement("tr");

            row.innerHTML = `
                <td>${id}</td>
                <td>${status}</td>
                <td>${size}</td>
                <td>${extension}</td>
                <td>${sha !== "-" ? sha.substring(0,16) + "..." : "-"}</td>
            `;

            table.appendChild(row);
        });

        document.getElementById("totalFiles").innerText = files.length;
        document.getElementById("completedCount").innerText = completed;
        document.getElementById("processingCount").innerText = processing;

    } catch (err) {
        console.error("Error loading files:", err);
    }
}

async function checkHealth() {
    try {
        const res = await fetch("/healthz");
        document.getElementById("healthStatus").innerText =
            res.ok ? "Healthy ✅" : "Unhealthy ❌";
    } catch (err) {
        document.getElementById("healthStatus").innerText = "Unhealthy ❌";
    }
}

async function init() {
    await loadFiles();
    await checkHealth();
    setInterval(loadFiles, 3000);
}

init();
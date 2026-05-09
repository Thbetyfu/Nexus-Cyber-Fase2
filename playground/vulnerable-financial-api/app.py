from flask import Flask, request, jsonify, render_template_string
import sqlite3
import os

app = Flask(__name__)
DB_NAME = "financial.db"

def init_db():
    conn = sqlite3.connect(DB_NAME)
    c = conn.cursor()
    c.execute("DROP TABLE IF EXISTS investors")
    c.execute('''CREATE TABLE investors (
                    id INTEGER PRIMARY KEY,
                    name TEXT,
                    account_number TEXT,
                    balance REAL,
                    tier TEXT,
                    secret_token TEXT
                 )''')
    data = [
        (1, "Bpk. Budi Santoso", "1001-8X9A", 550000000.00, "Platinum", "-"),
        (2, "Ibu Siti Aminah", "1002-4M2B", 12500000.00, "Gold", "-"),
        (3, "PT. Nusantara Jaya", "8001-9Z1C", 45000000000.00, "Corporate", "-"),
        (4, "SECRET_ADMIN_ACCOUNT", "0000-0000", 99999999999.00, "GOD_MODE", "ADMIN_OJK_KEY_XYZ890"),
        (5, "Dr. Hendra Wijaya", "1003-7L3D", 850000000.00, "Platinum", "-"),
    ]
    c.executemany("INSERT INTO investors VALUES (?, ?, ?, ?, ?, ?)", data)
    conn.commit()
    conn.close()

# Front-End Mockup: Financial Data Center Dashboard
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Financial Data Center | Otoritas Simulasi</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; color: #333; }
        .header { background-color: #0b3d91; color: #d4af37; padding: 20px 40px; border-bottom: 5px solid #d4af37; display: flex; justify-content: space-between; align-items: center; }
        .header h1 { margin: 0; font-size: 24px; letter-spacing: 1px; }
        .header span { font-size: 14px; color: #fff; }
        .container { max-width: 1000px; margin: 40px auto; background: #fff; padding: 30px; box-shadow: 0 4px 15px rgba(0,0,0,0.1); border-radius: 8px; }
        h2 { color: #0b3d91; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-top: 0; }
        .search-box { margin-bottom: 20px; display: flex; gap: 10px; }
        input[type="text"] { padding: 10px; width: 300px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; }
        button { background-color: #0b3d91; color: #fff; border: none; padding: 10px 20px; cursor: pointer; border-radius: 4px; font-size: 16px; font-weight: bold; }
        button:hover { background-color: #082963; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #eee; }
        th { background-color: #f8f9fa; color: #0b3d91; font-weight: 600; }
        tr:hover { background-color: #f1f5f9; }
        .tier-Platinum { color: #64748b; font-weight: bold; }
        .tier-Gold { color: #ca8a04; font-weight: bold; }
        .tier-Corporate { color: #0b3d91; font-weight: bold; }
        .footer { text-align: center; margin-top: 50px; color: #888; font-size: 12px; }
        #error-msg { color: #dc2626; font-weight: bold; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🏛️ Financial Data Center (Simulasi)</h1>
        <span>Sistem Informasi Nasabah Prioritas</span>
    </div>
    <div class="container">
        <h2>Pencarian Nasabah per-ID</h2>
        <div class="search-box">
            <input type="text" id="investorId" value="1" placeholder="Masukkan ID Nasabah (1-3)">
            <button onclick="fetchData()">Cari Nasabah</button>
        </div>
        <p class="text-xs text-gray-500">Mencari ID = <span id="query-display" style="font-family: monospace; background: #eee; padding: 2px 5px;">1</span></p>
        <div id="error-msg"></div>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Nama Nasabah</th>
                    <th>No. Rekening</th>
                    <th>Saldo (IDR)</th>
                    <th>Tier</th>
                </tr>
            </thead>
            <tbody id="table-body">
                <!-- Data injected via JS -->
            </tbody>
        </table>
    </div>
    <div class="footer">
        &copy; 2026 Simulasi Otoritas Finansial - Vulnerable by Design for Nexus Cyber Testing
    </div>

    <script>
        async function fetchData() {
            const id = document.getElementById('investorId').value;
            document.getElementById('query-display').innerText = id;
            try {
                const response = await fetch('/api/investors?id=' + encodeURIComponent(id));
                const resJson = await response.json();
                
                const tbody = document.getElementById('table-body');
                const errorDiv = document.getElementById('error-msg');
                tbody.innerHTML = '';
                errorDiv.innerText = '';

                if (resJson.status === 'success') {
                    resJson.data.forEach(row => {
                        const tr = document.createElement('tr');
                        tr.innerHTML = `
                            <td>${row.id}</td>
                            <td>${row.name}</td>
                            <td style="font-family: monospace;">${row.account}</td>
                            <td>Rp ${Number(row.balance).toLocaleString('id-ID')}</td>
                            <td class="tier-${row.tier}">${row.tier}</td>
                        `;
                        tbody.appendChild(tr);
                    });
                } else {
                    errorDiv.innerText = "Database Error: " + resJson.message;
                }
            } catch (err) {
                document.getElementById('error-msg').innerText = err.message;
            }
        }
        // Initial load
        fetchData();
    </script>
</body>
</html>
"""

@app.route("/")
def index():
    return render_template_string(HTML_TEMPLATE)

# VULNERABLE ENDPOINT - INTENTIONALLY RAW SQL
@app.route("/api/investors")
def get_investors():
    # VULNERABILITY: Mengambil payload langsung dari URL tanpa sanitasi (param: id)
    user_id = request.args.get("id", "1")
    
    conn = sqlite3.connect(DB_NAME)
    c = conn.cursor()
    
    # KEKURANGAN ARSITEKTURAL (Vulnerable-by-Design): String Concatenation murni
    query = f"SELECT id, name, account_number, balance, tier FROM investors WHERE id = {user_id}"
    
    try:
        c.execute(query)
        rows = c.fetchall()
        result = []
        for r in rows:
            result.append({
                "id": r[0], 
                "name": r[1], 
                "account": r[2], 
                "balance": r[3], 
                "tier": r[4]
            })
        return jsonify({"status": "success", "data": result, "query_executed": query})
    except Exception as e:
        # Leak detail DB ke attacker bila syntax salah
        return jsonify({"status": "error", "message": str(e), "query_executed": query})
    finally:
        conn.close()

if __name__ == "__main__":
    if not os.path.exists(DB_NAME):
        init_db()
    
    print("\n[!] WARNING: STARTING VULNERABLE INSTITUTION API")
    print("[!] Security mechanisms disabled. Endpoint /api/investors is vulnerable to pure SQLi.")
    # Kita bind pada port 3001 karena port 3000 sedang diokupasi oleh Nexus Dashboard (Tugas sebelumnya)
    app.run(host="0.0.0.0", port=3001)

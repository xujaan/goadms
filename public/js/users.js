(function() {
    function api(method, path, body) {
        return fetch(path, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: body ? JSON.stringify(body) : undefined
        }).then(function(r) { return r.json(); });
    }

    async function load() {
        try {
            var d = await api("GET", "/api/v1/fingerprint-users");
            var users = d.data || [];
            var h = "";
            users.forEach(function(u) {
                var nameEsc = u.full_name.replace(/'/g, "\\'");
                var deptEsc = (u.department || "").replace(/'/g, "\\'");
                h += '<div class="card bg-dark-subtle mb-1"><div class="card-body py-2"><div class="d-flex justify-content-between">' +
                    '<div><strong class="text-light">' + u.full_name + '</strong> ' +
                    '<code class="text-secondary">' + u.employee_code + '</code><br>' +
                    '<small class="text-secondary">' + (u.department || "—") + '</small></div>' +
                    '<div>' +
                    '<button class="btn btn-sm btn-outline-secondary me-1" onclick="Users.edit(\'' + u.id + '\',\'' + nameEsc + '\',\'' + deptEsc + '\',\'' + u.employee_code + '\')"><i class="bi bi-pencil"></i></button>' +
                    '<button class="btn btn-sm btn-outline-info me-1" onclick="Users.syncOne(\'' + u.id + '\')"><i class="bi bi-upload"></i></button>' +
                    '<button class="btn btn-sm btn-outline-danger" onclick="Users.del(\'' + u.id + '\')"><i class="bi bi-trash"></i></button>' +
                    '</div></div></div></div>';
            });
            document.getElementById("userList").innerHTML = h || '<div class="text-secondary">No users</div>';
        } catch (e) { console.error(e); }
    }

    window.Users = {
        load: load,
        add: async function(e) {
            e.preventDefault();
            var f = document.getElementById("userForm");
            var d = {};
            new FormData(f).forEach(function(v, k) { d[k] = v; });
            await api("POST", "/api/v1/fingerprint-users", d);
            f.reset();
            load();
        },
        edit: async function(id, name, dept, code) {
            var nn = prompt("Name", name);
            if (!nn) return;
            var dd = prompt("Department", dept);
            await api("PUT", "/api/v1/fingerprint-users/" + id, { full_name: nn, department: dd || "", employee_code: code });
            load();
        },
        del: async function(id) {
            if (!confirm("Delete user?")) return;
            await api("DELETE", "/api/v1/fingerprint-users/" + id);
            load();
        },
        syncOne: async function(id) {
            var did = document.getElementById("fuDevice").value;
            if (!did) { alert("Select a device first"); return; }
            var r = await api("POST", "/api/v1/fingerprint-users/" + id + "/sync", { device_id: did });
            alert(r.message || "Synced");
        },
        syncAll: async function() {
            var did = document.getElementById("fuDevice").value;
            if (!did) { alert("Select a device first"); return; }
            var r = await api("POST", "/api/v1/fingerprint-users/sync-all", { device_id: did });
            alert("Synced: " + (r.synced || 0) + " user(s)");
        }
    };

    load();
})();

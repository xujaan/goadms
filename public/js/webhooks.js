(function() {
    function api(method, path, body) {
        return fetch(path, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: body ? JSON.stringify(body) : undefined
        }).then(function(r) { return r.json(); });
    }

    function loadWebhooks() {
        var id = document.getElementById("whDevice").value;
        if (!id) {
            document.getElementById("whList").innerHTML = '<div class="text-secondary">Select a device</div>';
            return;
        }
        api("GET", "/api/v1/devices/" + id + "/webhooks").then(function(d) {
            var whs = d.data || [];
            var h = "<h5>Webhooks (" + whs.length + ")</h5>";
            whs.forEach(function(w) {
                var toggleBtn = w.is_active ?
                    '<button class="btn btn-sm btn-outline-warning me-1" onclick="Webhooks.toggle(\'' + w.id + '\',true)">Disable</button>' :
                    '<button class="btn btn-sm btn-outline-success me-1" onclick="Webhooks.toggle(\'' + w.id + '\',false)">Enable</button>';
                h += '<div class="card bg-dark-subtle mb-2" data-whid="' + w.id + '"><div class="card-body py-2"><div class="d-flex justify-content-between">' +
                    '<div><strong class="text-light wh-name">' + w.name + '</strong><br>' +
                    '<small class="text-secondary">URL: <span class="wh-url">' + w.url + '</span><br>Events: ' + (w.events||[]).join(", ") + '<br>' +
                    'Status: ' + (w.is_active ? '<span class="text-success">Active</span>' : '<span class="text-danger">Inactive</span>') +
                    '</small></div><div>' + toggleBtn +
                    '<button class="btn btn-sm btn-outline-secondary me-1" onclick="Webhooks.edit(\'' + w.id + '\')"><i class="bi bi-pencil"></i></button>' +
                    '<button class="btn btn-sm btn-outline-info me-1" onclick="Webhooks.test(\'' + w.id + '\')">Test</button>' +
                    '<button class="btn btn-sm btn-outline-danger" onclick="Webhooks.del(\'' + w.id + '\')">Del</button>' +
                    '</div></div></div></div>';
            });
            document.getElementById("whList").innerHTML = h || '<div class="text-secondary">No webhooks</div>';
        }).catch(function(e) { console.error(e); });
    }

    // Store webhook data for edit.
    var webhookCache = {};

    window.Webhooks = {
        load: loadWebhooks,
        toggle: async function(wid, active) {
            var did = document.getElementById("whDevice").value;
            await api("PUT", "/api/v1/devices/" + did + "/webhooks/" + wid, { is_active: !active });
            loadWebhooks();
        },
        edit: function(wid) {
            // Toggle inline edit form.
            var container = document.querySelector('[data-whid="' + wid + '"]');
            if (!container) return;
            var name = container.querySelector('.wh-name').textContent;
            var url = container.querySelector('.wh-url').textContent;
            container.innerHTML = '<div class="card-body py-2">' +
                '<input id="whEditName" class="form-control form-control-sm mb-1" value="' + name.replace(/"/g,'&quot;') + '">' +
                '<input id="whEditURL" class="form-control form-control-sm mb-1" value="' + url + '">' +
                '<button class="btn btn-sm btn-success me-1" onclick="Webhooks.saveEdit(\'' + wid + '\')">Save</button>' +
                '<button class="btn btn-sm btn-outline-secondary" onclick="Webhooks.load()">Cancel</button>' +
                '</div>';
        },
        saveEdit: async function(wid) {
            var did = document.getElementById("whDevice").value;
            var nn = document.getElementById("whEditName").value;
            var uu = document.getElementById("whEditURL").value;
            if (!nn || !uu) return;
            await api("PUT", "/api/v1/devices/" + did + "/webhooks/" + wid, { name: nn, url: uu });
            loadWebhooks();
        },
        test: async function(id) {
            var did = document.getElementById("whDevice").value;
            var r = await api("POST", "/api/v1/devices/" + did + "/webhooks/" + id + "/test");
            alert("Test: " + (r.message || "sent"));
        },
        del: async function(id) {
            if (!confirm("Delete webhook?")) return;
            var did = document.getElementById("whDevice").value;
            await api("DELETE", "/api/v1/devices/" + did + "/webhooks/" + id);
            loadWebhooks();
        },
        add: async function(e) {
            e.preventDefault();
            var f = document.getElementById("whForm");
            var id = document.getElementById("whDevice").value;
            if (!id) return;
            var d = {};
            new FormData(f).forEach(function(v, k) { d[k] = v; });
            d.events = Array.from(f.events.selectedOptions).map(function(o) { return o.value; });
            await api("POST", "/api/v1/devices/" + id + "/webhooks", d);
            f.reset();
            loadWebhooks();
        }
    };

    // Load device list into dropdown.
    fetch("/api/v1/devices").then(function(r) { return r.json(); }).then(function(d) {
        var sel = document.getElementById("whDevice");
        (d.data || []).forEach(function(d) {
            sel.innerHTML += '<option value="' + d.id + '">' + d.name + ' (' + (d.ip_address||"—") + ')</option>';
        });
        loadWebhooks();
    });
})();

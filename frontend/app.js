const state = {
  view: "home",
  mode: "login",
  token: localStorage.getItem("algojudge_token") || "",
  user: JSON.parse(localStorage.getItem("algojudge_user") || "null"),
  problems: [],
};

const nodes = {
  navLinks: [...document.querySelectorAll(".nav-link")],
  panels: [...document.querySelectorAll("[data-panel]")],
  authButton: document.getElementById("auth-button"),
  profileChip: document.getElementById("profile-chip"),
  authDialog: document.getElementById("auth-dialog"),
  uploadDialog: document.getElementById("upload-dialog"),
  authSubmit: document.getElementById("auth-submit"),
  authMessage: document.getElementById("auth-message"),
  authHint: document.getElementById("auth-hint"),
  authUsername: document.getElementById("auth-username"),
  authPassword: document.getElementById("auth-password"),
  loginTab: document.getElementById("login-tab"),
  registerTab: document.getElementById("register-tab"),
  uploadButton: document.getElementById("upload-button"),
  uploadSubmit: document.getElementById("upload-submit"),
  uploadMessage: document.getElementById("upload-message"),
  searchInput: document.getElementById("search-input"),
  feedbackPill: document.getElementById("feedback-pill"),
  problemTableBody: document.getElementById("problem-table-body"),
  problemDetail: document.getElementById("problem-detail"),
  statusProblems: document.getElementById("status-problems"),
  statusSubmissions: document.getElementById("status-submissions"),
  statusPending: document.getElementById("status-pending"),
  statusJudger: document.getElementById("status-judger"),
};

document.querySelectorAll("[data-view-target]").forEach((button) => {
  button.addEventListener("click", () => switchView(button.dataset.viewTarget));
});

nodes.navLinks.forEach((link) => {
  link.addEventListener("click", () => switchView(link.dataset.view));
});

document.getElementById("open-auth-inline").addEventListener("click", () => openAuth("login"));
nodes.authButton.addEventListener("click", () => {
  if (state.user) {
    logout();
    return;
  }
  openAuth("login");
});

nodes.loginTab.addEventListener("click", () => setAuthMode("login"));
nodes.registerTab.addEventListener("click", () => setAuthMode("register"));
nodes.authSubmit.addEventListener("click", submitAuth);
nodes.uploadButton.addEventListener("click", () => nodes.uploadDialog.showModal());
nodes.uploadSubmit.addEventListener("click", submitProblem);
nodes.searchInput.addEventListener("input", renderProblems);

switchView("home");
hydrateSession();
loadProblems();

function switchView(view) {
  state.view = view;
  nodes.navLinks.forEach((link) => link.classList.toggle("is-active", link.dataset.view === view));
  nodes.panels.forEach((panel) => panel.classList.toggle("hidden", panel.dataset.panel !== view));

  if (view === "problems") {
    loadProblems();
  }
  if (view === "status") {
    loadStatus();
  }
}

function setAuthMode(mode) {
  state.mode = mode;
  nodes.loginTab.classList.toggle("is-active", mode === "login");
  nodes.registerTab.classList.toggle("is-active", mode === "register");
  nodes.authSubmit.textContent = mode === "login" ? "Login" : "Register";
  nodes.authMessage.textContent = "";
}

function openAuth(mode) {
  setAuthMode(mode);
  nodes.authDialog.showModal();
}

async function submitAuth() {
  const payload = {
    username: nodes.authUsername.value.trim(),
    password: nodes.authPassword.value,
  };

  const route = state.mode === "login" ? "/api/v1/auth/login" : "/api/v1/auth/register";
  try {
    const response = await request(route, {
      method: "POST",
      body: JSON.stringify(payload),
    });

    persistSession(response.token, response.user);
    nodes.authMessage.textContent = state.mode === "login" ? "Logged in." : "Account created.";
    nodes.authMessage.className = "message success";
    nodes.authDialog.close();
    refreshAuthUI();
    loadProblems();
    if (state.user?.role === 1) {
      loadStatus();
    }
  } catch (error) {
    nodes.authMessage.textContent = error.message;
    nodes.authMessage.className = "message error";
  }
}

async function hydrateSession() {
  if (!state.token) {
    refreshAuthUI();
    return;
  }

  try {
    const response = await request("/api/v1/auth/me", {
      headers: authHeaders(),
    });
    state.user = response.user;
    localStorage.setItem("algojudge_user", JSON.stringify(state.user));
  } catch {
    logout(false);
  }
  refreshAuthUI();
}

function refreshAuthUI() {
  const isAdmin = state.user?.role === 1;
  nodes.profileChip.classList.toggle("hidden", !state.user);
  nodes.profileChip.textContent = state.user ? `${state.user.username} • ${isAdmin ? "Admin" : "Student"}` : "";
  nodes.authButton.textContent = state.user ? "Logout" : "Login";
  nodes.uploadButton.classList.toggle("hidden", !isAdmin);
  nodes.feedbackPill.textContent = isAdmin ? "Admin session active" : "Public problems only";
}

async function loadProblems() {
  try {
    const response = await request("/api/v1/problems");
    state.problems = response.items || [];
    renderProblems();
  } catch (error) {
    nodes.problemTableBody.innerHTML = `<tr><td colspan="5">${escapeHtml(error.message)}</td></tr>`;
  }
}

function renderProblems() {
  const keyword = nodes.searchInput.value.trim().toLowerCase();
  const filtered = state.problems.filter((problem) => problem.title.toLowerCase().includes(keyword));

  if (!filtered.length) {
    nodes.problemTableBody.innerHTML = `<tr><td colspan="5">No problems found.</td></tr>`;
    nodes.problemDetail.classList.add("hidden");
    return;
  }

  nodes.problemTableBody.innerHTML = filtered.map((problem) => `
    <tr data-problem-id="${problem.id}">
      <td>${problem.id}</td>
      <td>${escapeHtml(problem.title)}</td>
      <td>${problem.time_limit} ms</td>
      <td>${problem.memory_limit} MB</td>
      <td>${problem.status === 1 ? "Published" : "Pending"}</td>
    </tr>
  `).join("");

  [...nodes.problemTableBody.querySelectorAll("tr[data-problem-id]")].forEach((row) => {
    row.addEventListener("click", () => showProblemDetail(row.dataset.problemId));
  });
}

async function showProblemDetail(problemID) {
  try {
    const problem = await request(`/api/v1/problems/${problemID}`);
    nodes.problemDetail.classList.remove("hidden");
    nodes.problemDetail.innerHTML = `
      <p class="eyebrow">Problem #${problem.id}</p>
      <h3>${escapeHtml(problem.title)}</h3>
      <p>${escapeHtml(problem.description || "No description provided.")}</p>
      <div class="metric-grid">
        <div class="metric-card">
          <span class="metric-label">Time Limit</span>
          <strong>${problem.limits.time} ms</strong>
        </div>
        <div class="metric-card">
          <span class="metric-label">Memory Limit</span>
          <strong>${problem.limits.memory} MB</strong>
        </div>
      </div>
      ${(problem.samples || []).map((sample, index) => `
        <div class="sample-box">
          <strong>Sample ${index + 1}</strong>
          <p><b>Input:</b> ${escapeHtml(sample.input)}</p>
          <p><b>Output:</b> ${escapeHtml(sample.output)}</p>
        </div>
      `).join("")}
    `;
  } catch (error) {
    nodes.problemDetail.classList.remove("hidden");
    nodes.problemDetail.innerHTML = `<p>${escapeHtml(error.message)}</p>`;
  }
}

async function loadStatus() {
  if (state.user?.role !== 1) {
    nodes.statusProblems.textContent = "-";
    nodes.statusSubmissions.textContent = "-";
    nodes.statusPending.textContent = "-";
    nodes.statusJudger.textContent = "admin only";
    return;
  }

  try {
    const status = await request("/api/v1/admin/status", {
      headers: authHeaders(),
    });
    nodes.statusProblems.textContent = status.problem_count;
    nodes.statusSubmissions.textContent = status.submission_count;
    nodes.statusPending.textContent = status.pending_judges;
    nodes.statusJudger.textContent = status.judger_status;
  } catch (error) {
    nodes.statusJudger.textContent = error.message;
  }
}

async function submitProblem() {
  const payload = {
    title: document.getElementById("problem-title").value.trim(),
    description: document.getElementById("problem-description").value.trim(),
    time_limit: Number(document.getElementById("problem-time").value || 1000),
    memory_limit: Number(document.getElementById("problem-memory").value || 256),
    test_cases: [
      {
        input: document.getElementById("sample-input").value.trim(),
        output: document.getElementById("sample-output").value.trim(),
        is_sample: true,
      },
    ],
  };

  try {
    await request("/api/v1/admin/problems", {
      method: "POST",
      headers: authHeaders(),
      body: JSON.stringify(payload),
    });
    nodes.uploadMessage.textContent = "Problem published.";
    nodes.uploadMessage.className = "message success";
    nodes.uploadDialog.close();
    loadProblems();
  } catch (error) {
    nodes.uploadMessage.textContent = error.message;
    nodes.uploadMessage.className = "message error";
  }
}

function persistSession(token, user) {
  state.token = token;
  state.user = user;
  localStorage.setItem("algojudge_token", token);
  localStorage.setItem("algojudge_user", JSON.stringify(user));
}

function logout(updateUI = true) {
  state.token = "";
  state.user = null;
  localStorage.removeItem("algojudge_token");
  localStorage.removeItem("algojudge_user");
  if (updateUI) {
    refreshAuthUI();
  }
}

function authHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${state.token}`,
  };
}

async function request(url, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  const response = await fetch(url, { ...options, headers });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || "Request failed");
  }
  return data;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

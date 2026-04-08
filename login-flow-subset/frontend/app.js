const state = {
  mode: "login",
  token: localStorage.getItem("login_flow_subset_token") || "",
  user: JSON.parse(localStorage.getItem("login_flow_subset_user") || "null"),
  problems: [
    {
      id: 1001,
      title: "A + B Again",
      description: "A tiny problem kept here so the login demo still feels like an online judge page.",
    },
    {
      id: 1034,
      title: "Ribbon Intervals",
      description: "A mock medium problem card used only for frontend presentation.",
    },
    {
      id: 1112,
      title: "Matrix Echo",
      description: "Another mock problem card so the subset still looks like product UI.",
    },
  ],
};

const nodes = {
  sessionChip: document.getElementById("session-chip"),
  logoutButton: document.getElementById("logout-button"),
  authTitle: document.getElementById("auth-title"),
  loginTab: document.getElementById("login-tab"),
  registerTab: document.getElementById("register-tab"),
  authUsername: document.getElementById("auth-username"),
  authPassword: document.getElementById("auth-password"),
  authMessage: document.getElementById("auth-message"),
  authSubmit: document.getElementById("auth-submit"),
  sessionPanel: document.getElementById("session-panel"),
  problemList: document.getElementById("problem-list"),
};

bindEvents();
renderProblems();
refreshAuthUI();
hydrateSession();

function bindEvents() {
  nodes.loginTab.addEventListener("click", () => setMode("login"));
  nodes.registerTab.addEventListener("click", () => setMode("register"));
  nodes.authSubmit.addEventListener("click", submitAuth);
  nodes.logoutButton.addEventListener("click", logout);
}

function setMode(mode) {
  state.mode = mode;
  nodes.loginTab.classList.toggle("is-active", mode === "login");
  nodes.registerTab.classList.toggle("is-active", mode === "register");
  nodes.authTitle.textContent = mode === "login" ? "Login" : "Register";
  nodes.authSubmit.textContent = mode === "login" ? "Login" : "Register";
  nodes.authMessage.textContent = "";
  nodes.authMessage.className = "message";
}

async function submitAuth() {
  const payload = {
    username: nodes.authUsername.value.trim(),
    password: nodes.authPassword.value,
  };

  if (!payload.username || !payload.password) {
    setMessage("Username and password are required.", "error");
    return;
  }

  const route = state.mode === "login" ? "/api/v1/auth/login" : "/api/v1/auth/register";

  try {
    const response = await request(route, {
      method: "POST",
      body: JSON.stringify(payload),
    });
    persistSession(response.token, response.user);
    refreshAuthUI();
    setMessage(state.mode === "login" ? "Logged in." : "Account created.", "success");
  } catch (error) {
    setMessage(error.message, "error");
  }
}

async function hydrateSession() {
  if (!state.token) {
    return;
  }

  try {
    const response = await request("/api/v1/auth/me", {
      headers: authHeaders(),
    });
    state.user = response.user;
    localStorage.setItem("login_flow_subset_user", JSON.stringify(state.user));
  } catch {
    logout(false);
  }
  refreshAuthUI();
}

function refreshAuthUI() {
  nodes.logoutButton.classList.toggle("hidden", !state.user);
  nodes.sessionChip.textContent = state.user ? `${state.user.username} (${roleLabel(state.user.role)})` : "Signed out";

  if (!state.user) {
    nodes.sessionPanel.innerHTML = `<p class="copy">No active session.</p>`;
    return;
  }

  nodes.sessionPanel.innerHTML = `
    <div class="session-box">
      <h3>${escapeHtml(state.user.username)}</h3>
      <p>ID: ${state.user.id}</p>
      <p>Role: ${escapeHtml(roleLabel(state.user.role))}</p>
      <p>JWT saved in localStorage and verified with <code>/api/v1/auth/me</code>.</p>
    </div>
  `;
}

function renderProblems() {
  nodes.problemList.innerHTML = state.problems.map((problem) => `
    <article class="problem-card">
      <h3>${escapeHtml(problem.title)}</h3>
      <strong>#${problem.id}</strong>
      <p>${escapeHtml(problem.description)}</p>
    </article>
  `).join("");
}

function persistSession(token, user) {
  state.token = token;
  state.user = user;
  localStorage.setItem("login_flow_subset_token", token);
  localStorage.setItem("login_flow_subset_user", JSON.stringify(user));
}

function logout(updateUI = true) {
  state.token = "";
  state.user = null;
  localStorage.removeItem("login_flow_subset_token");
  localStorage.removeItem("login_flow_subset_user");
  if (updateUI) {
    refreshAuthUI();
    setMessage("Logged out.", "");
  }
}

function authHeaders() {
  return {
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

function roleLabel(role) {
  return Number(role) === 1 ? "Admin" : "Student";
}

function setMessage(message, type) {
  nodes.authMessage.textContent = message;
  nodes.authMessage.className = type ? `message ${type}` : "message";
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

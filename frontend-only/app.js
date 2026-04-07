const STORAGE_KEYS = {
  users: "algojudge_frontend_subset_users",
  session: "algojudge_frontend_subset_session",
  submissions: "algojudge_frontend_subset_submissions",
};

const DEMO_PROBLEMS = [
  {
    id: 1001,
    title: "A + B Again",
    difficulty: "Easy",
    description: "Read two integers and print their sum. This starter problem exists so the frontend can show the full browse-detail-submit loop.",
    time_limit: 1000,
    memory_limit: 256,
    samples: [{ input: "1 2", output: "3" }],
  },
  {
    id: 1034,
    title: "Ribbon Intervals",
    difficulty: "Medium",
    description: "Given colored ribbon segments, determine the minimum number of joins required to form a target interval.",
    time_limit: 1500,
    memory_limit: 512,
    samples: [{ input: "5\n1 4\n2 5\n", output: "2" }],
  },
  {
    id: 1112,
    title: "Matrix Echo",
    difficulty: "Hard",
    description: "Simulate row and column transforms, then report the final checksum of the matrix after all operations.",
    time_limit: 2000,
    memory_limit: 512,
    samples: [{ input: "3 3\n1 2 3\n...", output: "42" }],
  },
];

const ANNOUNCEMENTS = [
  { title: "Frontend subset extracted for standalone review.", date: "2026-04-07" },
  { title: "Mock auth and mock submissions now run in-browser.", date: "2026-04-07" },
  { title: "Problem browsing flow kept close to the original project UI.", date: "2026-04-07" },
];

const DEFAULT_USERS = [
  { username: "demo_student", password: "password", role: "Student" },
  { username: "demo_admin", password: "password", role: "Admin" },
];

const state = {
  view: "home",
  mode: "login",
  user: loadSession(),
  problems: DEMO_PROBLEMS,
  selectedProblemId: DEMO_PROBLEMS[0].id,
  submissions: loadSubmissions(),
};

const nodes = {
  navLinks: [...document.querySelectorAll(".nav-link")],
  panels: [...document.querySelectorAll("[data-panel]")],
  authButton: document.getElementById("auth-button"),
  modeChip: document.getElementById("mode-chip"),
  authDialog: document.getElementById("auth-dialog"),
  openAuthInline: document.getElementById("open-auth-inline"),
  closeAuth: document.getElementById("close-auth"),
  loginTab: document.getElementById("login-tab"),
  registerTab: document.getElementById("register-tab"),
  authTitle: document.getElementById("auth-title"),
  authSubmit: document.getElementById("auth-submit"),
  authUsername: document.getElementById("auth-username"),
  authPassword: document.getElementById("auth-password"),
  authAdmin: document.getElementById("auth-admin"),
  authMessage: document.getElementById("auth-message"),
  announcementList: document.getElementById("announcement-list"),
  problemSearch: document.getElementById("problem-search"),
  problemTableBody: document.getElementById("problem-table-body"),
  problemDetail: document.getElementById("problem-detail"),
  submissionTableBody: document.getElementById("submission-table-body"),
  statusProblemCount: document.getElementById("status-problem-count"),
  statusSubmissionCount: document.getElementById("status-submission-count"),
  statusAcceptedCount: document.getElementById("status-accepted-count"),
};

seedUsers();
bindEvents();
refreshAuthUI();
renderAnnouncements();
renderProblems();
renderProblemDetail();
renderSubmissions();

function bindEvents() {
  nodes.navLinks.forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.view));
  });

  document.querySelectorAll("[data-view-target]").forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.viewTarget));
  });

  nodes.authButton.addEventListener("click", () => {
    if (state.user) {
      logout();
      return;
    }
    openAuth("login");
  });

  nodes.openAuthInline.addEventListener("click", () => openAuth("login"));
  nodes.closeAuth.addEventListener("click", () => nodes.authDialog.close());
  nodes.loginTab.addEventListener("click", () => setAuthMode("login"));
  nodes.registerTab.addEventListener("click", () => setAuthMode("register"));
  nodes.authSubmit.addEventListener("click", submitAuth);
  nodes.problemSearch.addEventListener("input", renderProblems);
}

function switchView(view) {
  state.view = view;
  nodes.navLinks.forEach((button) => {
    button.classList.toggle("is-active", button.dataset.view === view);
  });
  nodes.panels.forEach((panel) => {
    panel.classList.toggle("hidden", panel.dataset.panel !== view);
  });
}

function setAuthMode(mode) {
  state.mode = mode;
  nodes.loginTab.classList.toggle("is-active", mode === "login");
  nodes.registerTab.classList.toggle("is-active", mode === "register");
  nodes.authTitle.textContent = mode === "login" ? "Login" : "Register";
  nodes.authSubmit.textContent = mode === "login" ? "Login" : "Register";
  nodes.authMessage.textContent = "";
}

function openAuth(mode) {
  setAuthMode(mode);
  nodes.authDialog.showModal();
}

function submitAuth() {
  const username = nodes.authUsername.value.trim();
  const password = nodes.authPassword.value;
  const wantsAdmin = nodes.authAdmin.checked;

  if (!username || !password) {
    showAuthMessage("Username and password are required.", "error");
    return;
  }

  const users = loadUsers();

  if (state.mode === "login") {
    const found = users.find((user) => user.username === username && user.password === password);
    if (!found) {
      showAuthMessage("Invalid demo credentials.", "error");
      return;
    }
    state.user = { username: found.username, role: found.role };
    persistSession();
    nodes.authDialog.close();
    refreshAuthUI();
    renderSubmissions();
    return;
  }

  if (users.some((user) => user.username === username)) {
    showAuthMessage("This demo username already exists.", "error");
    return;
  }

  users.push({
    username,
    password,
    role: wantsAdmin ? "Admin" : "Student",
  });
  localStorage.setItem(STORAGE_KEYS.users, JSON.stringify(users));
  state.user = { username, role: wantsAdmin ? "Admin" : "Student" };
  persistSession();
  nodes.authDialog.close();
  refreshAuthUI();
  renderSubmissions();
}

function refreshAuthUI() {
  if (state.user) {
    nodes.authButton.textContent = `${state.user.username} (${state.user.role}) / Logout`;
    nodes.modeChip.textContent = `${state.user.role} demo session`;
    return;
  }
  nodes.authButton.textContent = "Login";
  nodes.modeChip.textContent = "Local mock mode";
}

function renderAnnouncements() {
  nodes.announcementList.innerHTML = ANNOUNCEMENTS.map((item) => `
    <article class="announcement-item">
      <strong>${escapeHtml(item.title)}</strong>
      <p>${escapeHtml(item.date)}</p>
    </article>
  `).join("");
}

function renderProblems() {
  const query = nodes.problemSearch.value.trim().toLowerCase();
  const filtered = state.problems.filter((problem) => {
    if (!query) {
      return true;
    }
    return problem.title.toLowerCase().includes(query) || String(problem.id).includes(query);
  });

  if (!filtered.length) {
    nodes.problemTableBody.innerHTML = `<tr><td colspan="4" class="table-empty">No matching problems.</td></tr>`;
    return;
  }

  if (!filtered.some((problem) => problem.id === state.selectedProblemId)) {
    state.selectedProblemId = filtered[0].id;
  }

  nodes.problemTableBody.innerHTML = filtered.map((problem) => `
    <tr data-problem-id="${problem.id}" class="${problem.id === state.selectedProblemId ? "is-active" : ""}">
      <td>${problem.id}</td>
      <td>${escapeHtml(problem.title)}</td>
      <td>${escapeHtml(problem.difficulty)}</td>
      <td>${problem.time_limit} ms / ${problem.memory_limit} MB</td>
    </tr>
  `).join("");

  [...nodes.problemTableBody.querySelectorAll("[data-problem-id]")].forEach((row) => {
    row.addEventListener("click", () => {
      state.selectedProblemId = Number(row.dataset.problemId);
      renderProblems();
      renderProblemDetail();
    });
  });

  renderProblemDetail();
}

function renderProblemDetail() {
  const problem = state.problems.find((item) => item.id === state.selectedProblemId);
  if (!problem) {
    nodes.problemDetail.innerHTML = "<p class=\"table-empty\">Select a problem to view details.</p>";
    return;
  }

  nodes.problemDetail.innerHTML = `
    <div class="problem-detail-shell">
      <div>
        <h3 class="problem-title">${escapeHtml(problem.title)}</h3>
        <div class="problem-meta">
          <span>#${problem.id}</span>
          <span>${problem.time_limit} ms</span>
          <span>${problem.memory_limit} MB</span>
          <span class="difficulty-chip">${escapeHtml(problem.difficulty)}</span>
        </div>
      </div>

      <p class="problem-description">${escapeHtml(problem.description)}</p>

      <div class="sample-stack">
        ${problem.samples.map((sample, index) => `
          <div class="sample-box">
            <strong>Sample ${index + 1}</strong>
            <div>Input: ${escapeHtml(sample.input)}</div>
            <div>Output: ${escapeHtml(sample.output)}</div>
          </div>
        `).join("")}
      </div>

      <section class="composer-card">
        <div>
          <p class="eyebrow">Mock submission</p>
          <h3 class="problem-title">Submit in frontend-only mode</h3>
          <p class="submission-note">This does not call any backend endpoint. Verdicts are generated locally so the UI can be demonstrated in isolation.</p>
        </div>

        <div class="composer-toolbar">
          <select id="submit-language">
            <option value="cpp">C++</option>
            <option value="java">Java</option>
            <option value="python">Python</option>
          </select>
        </div>

        <label class="field">
          <span>Code</span>
          <textarea id="submit-code" spellcheck="false" placeholder="// Write your solution here">${escapeHtml(defaultCode())}</textarea>
        </label>

        <p id="submit-message" class="message"></p>
        <button id="submit-button" class="primary-button wide" type="button">Run Mock Submission</button>
      </section>
    </div>
  `;

  document.getElementById("submit-button").addEventListener("click", () => submitSolution(problem));
}

function submitSolution(problem) {
  if (!state.user) {
    openAuth("login");
    showSubmitMessage("Login is required for the demo submission flow.", "error");
    return;
  }

  const language = document.getElementById("submit-language").value;
  const code = document.getElementById("submit-code").value.trim();
  if (!code) {
    showSubmitMessage("Please enter some code first.", "error");
    return;
  }

  showSubmitMessage("Running locally...", "");

  window.setTimeout(() => {
    const verdict = selectVerdict(code);
    const submission = {
      id: Date.now(),
      problem_id: problem.id,
      problem_title: problem.title,
      language,
      status: verdict,
      runtime: 24 + (code.length % 180),
      memory: 2048 + ((problem.id + code.length) % 4096),
      created_at: new Date().toISOString(),
      username: state.user.username,
    };

    state.submissions = [submission, ...state.submissions].slice(0, 12);
    persistSubmissions();
    renderSubmissions();
    showSubmitMessage(`Submission finished with ${verdict}.`, verdict === "AC" ? "success" : "error");
    switchView("status");
  }, 720);
}

function renderSubmissions() {
  const visible = state.submissions.filter((item) => !state.user || item.username === state.user.username);
  nodes.statusProblemCount.textContent = String(state.problems.length);
  nodes.statusSubmissionCount.textContent = String(visible.length);
  nodes.statusAcceptedCount.textContent = String(visible.filter((item) => item.status === "AC").length);

  if (!visible.length) {
    nodes.submissionTableBody.innerHTML = `<tr><td colspan="7" class="table-empty">No submissions yet. Open a problem and run a mock submission.</td></tr>`;
    return;
  }

  nodes.submissionTableBody.innerHTML = visible.map((item) => `
    <tr>
      <td>${item.id}</td>
      <td>${escapeHtml(item.problem_title)}</td>
      <td>${escapeHtml(item.language)}</td>
      <td><span class="verdict-chip verdict-${String(item.status).toLowerCase()}">${escapeHtml(item.status)}</span></td>
      <td>${item.runtime} ms</td>
      <td>${item.memory} KB</td>
      <td>${new Date(item.created_at).toLocaleString()}</td>
    </tr>
  `).join("");
}

function showAuthMessage(message, type) {
  nodes.authMessage.textContent = message;
  nodes.authMessage.className = type ? `message ${type}` : "message";
}

function showSubmitMessage(message, type) {
  const node = document.getElementById("submit-message");
  if (!node) {
    return;
  }
  node.textContent = message;
  node.className = type ? `message ${type}` : "message";
}

function logout() {
  state.user = null;
  localStorage.removeItem(STORAGE_KEYS.session);
  refreshAuthUI();
  renderSubmissions();
  switchView("home");
}

function loadUsers() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.users) || "[]");
  } catch {
    return [];
  }
}

function seedUsers() {
  if (loadUsers().length) {
    return;
  }
  localStorage.setItem(STORAGE_KEYS.users, JSON.stringify(DEFAULT_USERS));
}

function loadSession() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.session) || "null");
  } catch {
    return null;
  }
}

function persistSession() {
  localStorage.setItem(STORAGE_KEYS.session, JSON.stringify(state.user));
}

function loadSubmissions() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.submissions) || "[]");
  } catch {
    return [];
  }
}

function persistSubmissions() {
  localStorage.setItem(STORAGE_KEYS.submissions, JSON.stringify(state.submissions));
}

function selectVerdict(code) {
  const options = ["AC", "WA", "TLE", "RE"];
  return options[code.length % options.length];
}

function defaultCode() {
  return [
    "#include <bits/stdc++.h>",
    "using namespace std;",
    "",
    "int main() {",
    "    ios::sync_with_stdio(false);",
    "    cin.tie(nullptr);",
    "    return 0;",
    "}",
  ].join("\n");
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

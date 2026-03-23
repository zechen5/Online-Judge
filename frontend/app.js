const state = {
  view: "home",
  mode: "login",
  token: localStorage.getItem("algojudge_token") || "",
  user: JSON.parse(localStorage.getItem("algojudge_user") || "null"),
  problems: [],
  adminProblems: [],
  submissions: [],
  uploadCases: [createTestCase(true)],
  editorCases: [],
  editorPrefs: loadEditorPrefs(),
  editorDrafts: loadEditorDrafts(),
  selectedProblem: null,
  announcements: [
    { title: "Submission history is now visible from the Status page.", date: "2026-03-22" },
    { title: "Authenticated users can submit code directly from problem details.", date: "2026-03-22" },
    { title: "Admin upload flow is available from the problem board.", date: "2026-03-21" },
    { title: "JWT login and role-based access control are enabled.", date: "2026-03-20" },
  ],
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
  authUsername: document.getElementById("auth-username"),
  authPassword: document.getElementById("auth-password"),
  loginTab: document.getElementById("login-tab"),
  registerTab: document.getElementById("register-tab"),
  uploadButton: document.getElementById("upload-button"),
  uploadSubmit: document.getElementById("upload-submit"),
  uploadMessage: document.getElementById("upload-message"),
  addUploadCaseButton: document.getElementById("add-upload-case"),
  uploadCaseList: document.getElementById("upload-case-list"),
  searchInput: document.getElementById("search-input"),
  homeSearchInput: document.getElementById("home-search-input"),
  homeSearchButton: document.getElementById("home-search-button"),
  feedbackPill: document.getElementById("feedback-pill"),
  adminReviewSection: document.getElementById("admin-review-section"),
  adminReviewBody: document.getElementById("admin-review-body"),
  adminPendingCount: document.getElementById("admin-pending-count"),
  adminPublishedCount: document.getElementById("admin-published-count"),
  adminHiddenCount: document.getElementById("admin-hidden-count"),
  problemTableBody: document.getElementById("problem-table-body"),
  problemDetail: document.getElementById("problem-detail"),
  announcementBody: document.getElementById("announcement-body"),
  latestProblemsBody: document.getElementById("latest-problems-body"),
  statusProblems: document.getElementById("status-problems"),
  statusSubmissions: document.getElementById("status-submissions"),
  statusPending: document.getElementById("status-pending"),
  statusJudger: document.getElementById("status-judger"),
  statusHint: document.getElementById("status-hint"),
  submissionTableBody: document.getElementById("submission-table-body"),
};

document.querySelectorAll("[data-view-target]").forEach((button) => {
  button.addEventListener("click", () => switchView(button.dataset.viewTarget));
});

nodes.navLinks.forEach((link) => {
  link.addEventListener("click", () => switchView(link.dataset.view));
});

document.getElementById("open-auth-inline").addEventListener("click", () => openAuth("login"));
document.getElementById("open-auth-side").addEventListener("click", () => openAuth("login"));
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
nodes.uploadButton.addEventListener("click", openUploadDialog);
nodes.uploadSubmit.addEventListener("click", submitProblem);
nodes.addUploadCaseButton.addEventListener("click", () => {
  state.uploadCases.push(createTestCase(false));
  renderUploadCases();
});
nodes.searchInput.addEventListener("input", renderProblems);
nodes.homeSearchButton.addEventListener("click", () => runQuickSearch(nodes.homeSearchInput.value));
nodes.homeSearchInput.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    runQuickSearch(nodes.homeSearchInput.value);
  }
});

switchView("home");
hydrateSession();
loadProblems();
renderHomePanels();

function switchView(view) {
  state.view = view;
  nodes.navLinks.forEach((link) => link.classList.toggle("is-active", link.dataset.view === view));
  nodes.panels.forEach((panel) => panel.classList.toggle("hidden", panel.dataset.panel !== view));

  if (view === "problems") {
    loadProblems();
    if (state.user?.role === 1) {
      loadAdminProblems();
    }
  }
  if (view === "status") {
    loadStatus();
    loadRecentSubmissions();
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

function openUploadDialog() {
  state.uploadCases = [createTestCase(true)];
  document.getElementById("problem-title").value = "";
  document.getElementById("problem-description").value = "";
  document.getElementById("problem-time").value = "1000";
  document.getElementById("problem-memory").value = "256";
  nodes.uploadMessage.textContent = "";
  renderUploadCases();
  nodes.uploadDialog.showModal();
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
    loadAdminProblems();
    loadStatus();
    loadRecentSubmissions();
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
  nodes.profileChip.textContent = state.user ? `${state.user.username} - ${isAdmin ? "Admin" : "Student"}` : "";
  nodes.authButton.textContent = state.user ? "Logout" : "Login";
  nodes.uploadButton.classList.toggle("hidden", !isAdmin);
  nodes.adminReviewSection.classList.toggle("hidden", !isAdmin);
  nodes.feedbackPill.textContent = isAdmin ? "Admin session active" : "Public problems only";
  nodes.statusHint.textContent = state.user ? "Signed in" : "Login required";
  if (!isAdmin) {
    state.adminProblems = [];
    renderAdminReview();
  }
}

async function loadProblems() {
  try {
    const response = await request("/api/v1/problems");
    state.problems = response.items || [];
    renderProblems();
    renderHomePanels();
  } catch (error) {
    nodes.problemTableBody.innerHTML = `<tr><td colspan="5">${escapeHtml(error.message)}</td></tr>`;
    renderHomePanels();
  }
}

async function loadAdminProblems() {
  if (state.user?.role !== 1) {
    return;
  }

  try {
    const response = await request("/api/v1/admin/problems?page_size=100", {
      headers: authHeaders(),
    });
    state.adminProblems = response.items || [];
    renderAdminReview();
  } catch (error) {
    nodes.adminReviewBody.innerHTML = `<tr><td colspan="5">${escapeHtml(error.message)}</td></tr>`;
  }
}

function renderHomePanels() {
  renderAnnouncements();
  renderLatestProblems();
}

function renderUploadCases() {
  renderCaseEditor("upload", nodes.uploadCaseList, state.uploadCases);
}

function renderEditorCases() {
  const container = document.getElementById("admin-case-list");
  if (!container) {
    return;
  }
  renderCaseEditor("admin", container, state.editorCases);
}

function renderCaseEditor(mode, container, cases) {
  container.innerHTML = cases.map((testCase, index) => `
    <article class="case-card">
      <div class="case-card-header">
        <strong>Testcase ${index + 1}</strong>
        <div class="table-actions">
          <label class="case-toggle">
            <input type="checkbox" data-case-role="${mode}" data-case-field="is_sample" data-case-index="${index}" ${testCase.is_sample ? "checked" : ""}>
            <span>Sample</span>
          </label>
          <button class="ghost-button mini-button" data-case-role="${mode}" data-case-action="remove" data-case-index="${index}" type="button" ${cases.length === 1 ? "disabled" : ""}>Remove</button>
        </div>
      </div>
      <div class="grid-two">
        <label class="field">
          <span>Input</span>
          <textarea data-case-role="${mode}" data-case-field="input" data-case-index="${index}" rows="4">${escapeHtml(testCase.input)}</textarea>
        </label>
        <label class="field">
          <span>Output</span>
          <textarea data-case-role="${mode}" data-case-field="output" data-case-index="${index}" rows="4">${escapeHtml(testCase.output)}</textarea>
        </label>
      </div>
    </article>
  `).join("");

  [...container.querySelectorAll(`[data-case-role="${mode}"][data-case-field]`)].forEach((element) => {
    const eventName = element.type === "checkbox" ? "change" : "input";
    element.addEventListener(eventName, (event) => {
      const field = event.currentTarget.dataset.caseField;
      const index = Number(event.currentTarget.dataset.caseIndex);
      const value = field === "is_sample" ? event.currentTarget.checked : event.currentTarget.value;
      updateCaseState(mode, index, field, value);
    });
  });

  [...container.querySelectorAll(`[data-case-role="${mode}"][data-case-action="remove"]`)].forEach((button) => {
    button.addEventListener("click", () => {
      removeCase(mode, Number(button.dataset.caseIndex));
    });
  });
}

function renderAdminReview() {
  const pendingCount = state.adminProblems.filter((problem) => problem.status === 0).length;
  const publishedCount = state.adminProblems.filter((problem) => problem.status === 1).length;
  const hiddenCount = state.adminProblems.filter((problem) => problem.status === 2).length;

  nodes.adminPendingCount.textContent = pendingCount;
  nodes.adminPublishedCount.textContent = publishedCount;
  nodes.adminHiddenCount.textContent = hiddenCount;

  if (state.user?.role !== 1) {
    nodes.adminReviewBody.innerHTML = "";
    return;
  }

  if (!state.adminProblems.length) {
    nodes.adminReviewBody.innerHTML = `<tr><td colspan="5">No problems in the review queue.</td></tr>`;
    return;
  }

  nodes.adminReviewBody.innerHTML = state.adminProblems.map((problem) => `
    <tr data-admin-problem-id="${problem.id}">
      <td>${problem.id}</td>
      <td>${escapeHtml(problem.title)}</td>
      <td><span class="status-badge status-${statusClassName(problem.status)}">${escapeHtml(statusLabel(problem.status))}</span></td>
      <td>${formatDate(problem.updated_at)}</td>
      <td>
        <div class="table-actions">
          ${problem.status !== 1 ? `<button class="ghost-button mini-button" data-admin-action="publish" data-problem-id="${problem.id}" type="button">Publish</button>` : ""}
          ${problem.status !== 2 ? `<button class="ghost-button mini-button" data-admin-action="hide" data-problem-id="${problem.id}" type="button">Hide</button>` : ""}
          ${problem.status !== 0 ? `<button class="ghost-button mini-button" data-admin-action="pending" data-problem-id="${problem.id}" type="button">Pending</button>` : ""}
          <button class="ghost-button mini-button" data-admin-action="edit" data-problem-id="${problem.id}" type="button">Edit</button>
          <button class="ghost-button mini-button danger-button" data-admin-action="delete" data-problem-id="${problem.id}" type="button">Delete</button>
        </div>
      </td>
    </tr>
  `).join("");

  [...nodes.adminReviewBody.querySelectorAll("tr[data-admin-problem-id]")].forEach((row) => {
    row.addEventListener("click", () => showProblemDetail(row.dataset.adminProblemId, true));
  });

  [...nodes.adminReviewBody.querySelectorAll("button[data-admin-action]")].forEach((button) => {
    button.addEventListener("click", async (event) => {
      event.stopPropagation();
      const problemID = Number(button.dataset.problemId);
      switch (button.dataset.adminAction) {
      case "publish":
        await updateProblemStatus(problemID, 1);
        break;
      case "hide":
        await updateProblemStatus(problemID, 2);
        break;
      case "pending":
        await updateProblemStatus(problemID, 0);
        break;
      case "edit":
        await showProblemDetail(problemID, true);
        break;
      case "delete":
        await deleteProblem(problemID);
        break;
      default:
        break;
      }
    });
  });
}

function renderAnnouncements() {
  nodes.announcementBody.innerHTML = state.announcements.map((item) => `
    <tr>
      <td>${escapeHtml(item.title)}</td>
      <td>${escapeHtml(item.date)}</td>
    </tr>
  `).join("");
}

function renderLatestProblems() {
  const latest = [...state.problems].slice(0, 8);
  if (!latest.length) {
    nodes.latestProblemsBody.innerHTML = `<tr><td colspan="3">No published problems yet.</td></tr>`;
    return;
  }

  nodes.latestProblemsBody.innerHTML = latest.map((problem) => `
    <tr data-home-problem-id="${problem.id}">
      <td>${escapeHtml(problem.title)}</td>
      <td>${problem.time_limit} ms / ${problem.memory_limit} MB</td>
      <td>${escapeHtml(statusLabel(problem.status))}</td>
    </tr>
  `).join("");

  [...nodes.latestProblemsBody.querySelectorAll("tr[data-home-problem-id]")].forEach((row) => {
    row.addEventListener("click", () => {
      switchView("problems");
      showProblemDetail(row.dataset.homeProblemId);
    });
  });
}

function renderProblems() {
  const filtered = filterProblems(nodes.searchInput.value);

  if (!filtered.length) {
    nodes.problemTableBody.innerHTML = `<tr><td colspan="5">No problems found.</td></tr>`;
    nodes.problemDetail.classList.add("hidden");
    return filtered;
  }

  nodes.problemTableBody.innerHTML = filtered.map((problem) => `
    <tr data-problem-id="${problem.id}">
      <td>${problem.id}</td>
      <td>${escapeHtml(problem.title)}</td>
      <td>${problem.time_limit} ms</td>
      <td>${problem.memory_limit} MB</td>
      <td>${escapeHtml(statusLabel(problem.status))}</td>
    </tr>
  `).join("");

  [...nodes.problemTableBody.querySelectorAll("tr[data-problem-id]")].forEach((row) => {
    row.addEventListener("click", () => showProblemDetail(row.dataset.problemId));
  });

  return filtered;
}

async function runQuickSearch(rawQuery) {
  const query = rawQuery.trim();
  nodes.searchInput.value = query;
  switchView("problems");

  if (!state.problems.length) {
    await loadProblems();
  }

  const matches = renderProblems();
  if (!query) {
    return;
  }

  const exactIdMatch = state.problems.find((problem) => String(problem.id) === query);
  if (exactIdMatch) {
    showProblemDetail(exactIdMatch.id);
    return;
  }

  if (matches.length === 1) {
    showProblemDetail(matches[0].id);
  }
}

function filterProblems(rawQuery) {
  const query = rawQuery.trim().toLowerCase();
  if (!query) {
    return state.problems;
  }

  return state.problems.filter((problem) => {
    const title = problem.title.toLowerCase();
    const id = String(problem.id);
    return title.includes(query) || id.includes(query);
  });
}

async function showProblemDetail(problemID, forceAdmin = false) {
  const useAdminRoute = forceAdmin || state.user?.role === 1;
  const route = useAdminRoute ? `/api/v1/admin/problems/${problemID}` : `/api/v1/problems/${problemID}`;
  try {
    const problem = normalizeProblem(await request(route, useAdminRoute ? { headers: authHeaders() } : {}));
    state.selectedProblem = problem;
    renderProblemDetail(problem);
  } catch (error) {
    nodes.problemDetail.classList.remove("hidden");
    nodes.problemDetail.innerHTML = `<p>${escapeHtml(error.message)}</p>`;
  }
}

function renderProblemDetail(problem) {
  const isAdmin = state.user?.role === 1;
  const allCases = problem.test_cases?.length ? problem.test_cases : (problem.samples || []).map((sample) => ({
    input: sample.input,
    output: sample.output,
    is_sample: true,
  }));
  const samples = allCases.filter((sample) => sample.is_sample !== false);
  const submitDisabled = problem.status !== 1;

  if (isAdmin) {
    state.editorCases = allCases.length ? allCases.map((item) => ({
      input: item.input || "",
      output: item.output || "",
      is_sample: item.is_sample !== false,
    })) : [createTestCase(true)];
  }

  nodes.problemDetail.classList.remove("hidden");
  nodes.problemDetail.innerHTML = `
    <div class="detail-grid">
      <article class="detail-card">
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
        ${(samples || []).map((sample, index) => `
          <div class="sample-box">
            <strong>Sample ${index + 1}</strong>
            <p><b>Input:</b> ${escapeHtml(sample.input)}</p>
            <p><b>Output:</b> ${escapeHtml(sample.output)}</p>
          </div>
        `).join("")}
      </article>

      <aside class="detail-side">
        ${isAdmin ? `
          <section class="detail-card submit-panel">
            <p class="eyebrow">Admin Review</p>
            <h3>Moderate problem</h3>
            <div class="grid-two">
              <label class="field">
                <span>Title</span>
                <input id="admin-problem-title" type="text" value="${escapeHtml(problem.title)}">
              </label>
              <label class="field">
                <span>Status</span>
                <select id="admin-problem-status" class="select-input">
                  <option value="0" ${Number(problem.status) === 0 ? "selected" : ""}>Pending</option>
                  <option value="1" ${Number(problem.status) === 1 ? "selected" : ""}>Published</option>
                  <option value="2" ${Number(problem.status) === 2 ? "selected" : ""}>Hidden</option>
                </select>
              </label>
            </div>
            <div class="grid-two">
              <label class="field">
                <span>Time Limit (ms)</span>
                <input id="admin-problem-time" type="number" value="${problem.time_limit || problem.limits.time}">
              </label>
              <label class="field">
                <span>Memory Limit (MB)</span>
                <input id="admin-problem-memory" type="number" value="${problem.memory_limit || problem.limits.memory}">
              </label>
            </div>
            <label class="field">
              <span>Description</span>
              <textarea id="admin-problem-description" rows="6">${escapeHtml(problem.description || "")}</textarea>
            </label>
            <section class="field">
              <div class="field-header">
                <span>Testcases</span>
                <button id="admin-add-case" class="ghost-button small-button" type="button">Add Testcase</button>
              </div>
              <div id="admin-case-list" class="case-list"></div>
            </section>
            <p id="admin-problem-message" class="message"></p>
            <div class="table-actions detail-actions">
              <button id="admin-save-problem" class="primary-button small-button" type="button">Save</button>
              <button id="admin-publish-problem" class="ghost-button small-button" type="button">Publish</button>
              <button id="admin-hide-problem" class="ghost-button small-button" type="button">Hide</button>
              <button id="admin-delete-problem" class="ghost-button small-button danger-button" type="button">Delete</button>
            </div>
          </section>
        ` : ""}

        <section class="detail-card submit-panel">
          <p class="eyebrow">Submit</p>
          <h3>Code submission</h3>
          ${submitDisabled ? `<p class="subtle-copy">Only published problems accept submissions.</p>` : ""}
          <div class="editor-toolbar">
            <label class="field compact-field">
              <span>Language</span>
              <select id="submit-language" class="select-input" ${submitDisabled ? "disabled" : ""}>
                <option value="cpp">C++</option>
                <option value="java">Java</option>
                <option value="python">Python</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>Indent</span>
              <select id="editor-indent-size" class="select-input" ${submitDisabled ? "disabled" : ""}>
                <option value="2" ${state.editorPrefs.indentSize === 2 ? "selected" : ""}>2 spaces</option>
                <option value="4" ${state.editorPrefs.indentSize === 4 ? "selected" : ""}>4 spaces</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>Font</span>
              <select id="editor-font-size" class="select-input" ${submitDisabled ? "disabled" : ""}>
                <option value="13" ${state.editorPrefs.fontSize === 13 ? "selected" : ""}>13 px</option>
                <option value="14" ${state.editorPrefs.fontSize === 14 ? "selected" : ""}>14 px</option>
                <option value="16" ${state.editorPrefs.fontSize === 16 ? "selected" : ""}>16 px</option>
              </select>
            </label>
            <label class="editor-toggle">
              <input id="editor-wrap-lines" type="checkbox" ${state.editorPrefs.wrap ? "checked" : ""} ${submitDisabled ? "disabled" : ""}>
              <span>Wrap</span>
            </label>
            <button id="editor-format" class="ghost-button small-button" type="button" ${submitDisabled ? "disabled" : ""}>Format</button>
          </div>
          <div id="code-editor" class="code-editor ${state.editorPrefs.wrap ? "is-wrap" : ""}">
            <div id="editor-gutter" class="editor-gutter">1</div>
            <div class="editor-surface">
              <pre id="editor-highlight" class="editor-highlight"><code></code></pre>
              <textarea id="submit-code" class="code-input code-editor-input" rows="14" spellcheck="false" placeholder="// Write your solution here" ${submitDisabled ? "disabled" : ""}></textarea>
              <div id="editor-autocomplete" class="editor-autocomplete hidden"></div>
            </div>
          </div>
          <p class="editor-hint">Tab and Shift+Tab control indentation. Enter keeps current indent.</p>
          <p id="submit-message" class="message"></p>
          <button id="submit-solution" class="primary-button wide" type="button" ${submitDisabled ? "disabled" : ""}>Run Submission</button>
          <div id="submission-result" class="result-card hidden"></div>
        </section>
      </aside>
    </div>
  `;

  initSubmissionEditor();

  if (!submitDisabled) {
    document.getElementById("submit-solution").addEventListener("click", () => submitSolution(problem.id));
  }

  if (isAdmin) {
    renderEditorCases();
    document.getElementById("admin-add-case").addEventListener("click", () => {
      state.editorCases.push(createTestCase(false));
      renderEditorCases();
    });
    document.getElementById("admin-save-problem").addEventListener("click", () => saveProblem(problem.id));
    document.getElementById("admin-publish-problem").addEventListener("click", () => updateProblemStatus(problem.id, 1, true));
    document.getElementById("admin-hide-problem").addEventListener("click", () => updateProblemStatus(problem.id, 2, true));
    document.getElementById("admin-delete-problem").addEventListener("click", () => deleteProblem(problem.id));
  }
}

async function submitSolution(problemID) {
  if (!state.user) {
    openAuth("login");
    return;
  }

  const message = document.getElementById("submit-message");
  const resultBox = document.getElementById("submission-result");
  const payload = {
    problem_id: Number(problemID),
    language: document.getElementById("submit-language").value,
    code: document.getElementById("submit-code").value,
  };

  try {
    const submission = await request("/api/v1/submissions", {
      method: "POST",
      headers: authHeaders(),
      body: JSON.stringify(payload),
    });

    message.textContent = "Submission queued. Polling for verdict...";
    message.className = "message success";
    resultBox.classList.remove("hidden");
    renderSubmissionResult(submission);
    loadStatus();
    loadRecentSubmissions();
    trackSubmission(submission.id);
  } catch (error) {
    message.textContent = error.message;
    message.className = "message error";
    resultBox.classList.add("hidden");
  }
}

async function loadStatus() {
  if (state.user?.role !== 1) {
    nodes.statusProblems.textContent = state.problems.length || "-";
    nodes.statusPending.textContent = "-";
    nodes.statusJudger.textContent = "async-local";
  }

  try {
    if (state.user?.role === 1) {
      const status = await request("/api/v1/admin/status", {
        headers: authHeaders(),
      });
      nodes.statusProblems.textContent = status.problem_count;
      nodes.statusSubmissions.textContent = status.submission_count;
      nodes.statusPending.textContent = status.pending_judges;
      nodes.statusJudger.textContent = status.judger_status;
      return;
    }

    nodes.statusProblems.textContent = state.problems.length || 0;
    nodes.statusSubmissions.textContent = state.submissions.length || 0;
    nodes.statusPending.textContent = state.submissions.filter((item) => isSubmissionActive(item.status)).length;
    nodes.statusJudger.textContent = "async-local";
  } catch (error) {
    nodes.statusJudger.textContent = error.message;
  }
}

async function loadRecentSubmissions() {
  if (!state.user) {
    state.submissions = [];
    renderRecentSubmissions();
    loadStatus();
    return;
  }

  try {
    const response = await request("/api/v1/submissions", {
      headers: authHeaders(),
    });
    state.submissions = response.items || [];
    renderRecentSubmissions();
    if (state.user?.role !== 1) {
      loadStatus();
    }
  } catch (error) {
    nodes.submissionTableBody.innerHTML = `<tr><td colspan="7">${escapeHtml(error.message)}</td></tr>`;
  }
}

function renderRecentSubmissions() {
  if (!state.user) {
    nodes.submissionTableBody.innerHTML = `<tr><td colspan="7">Login to view your submissions.</td></tr>`;
    return;
  }

  if (!state.submissions.length) {
    nodes.submissionTableBody.innerHTML = `<tr><td colspan="7">No submissions yet.</td></tr>`;
    return;
  }

  nodes.submissionTableBody.innerHTML = state.submissions.map((item) => `
    <tr>
      <td>${item.id}</td>
      <td>${escapeHtml(item.problem_title || `Problem #${item.problem_id}`)}</td>
      <td>${escapeHtml(item.language)}</td>
      <td><span class="status-badge status-${String(item.status).toLowerCase()}">${escapeHtml(item.status)}</span></td>
      <td>${item.runtime} ms</td>
      <td>${item.memory} KB</td>
      <td>${formatDate(item.created_at)}</td>
    </tr>
  `).join("");
}

async function submitProblem() {
  const payload = {
    title: document.getElementById("problem-title").value.trim(),
    description: document.getElementById("problem-description").value.trim(),
    time_limit: Number(document.getElementById("problem-time").value || 1000),
    memory_limit: Number(document.getElementById("problem-memory").value || 256),
    test_cases: state.uploadCases.map((testCase) => ({
      input: testCase.input.trim(),
      output: testCase.output.trim(),
      is_sample: Boolean(testCase.is_sample),
    })),
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
    loadAdminProblems();
  } catch (error) {
    nodes.uploadMessage.textContent = error.message;
    nodes.uploadMessage.className = "message error";
  }
}

async function saveProblem(problemID) {
  const message = document.getElementById("admin-problem-message");
  const payload = {
    title: document.getElementById("admin-problem-title").value.trim(),
    description: document.getElementById("admin-problem-description").value.trim(),
    time_limit: Number(document.getElementById("admin-problem-time").value || 1000),
    memory_limit: Number(document.getElementById("admin-problem-memory").value || 256),
    status: Number(document.getElementById("admin-problem-status").value),
    test_cases: state.editorCases.map((testCase) => ({
      input: testCase.input.trim(),
      output: testCase.output.trim(),
      is_sample: Boolean(testCase.is_sample),
    })),
  };

  try {
    await request(`/api/v1/admin/problems/${problemID}`, {
      method: "PUT",
      headers: authHeaders(),
      body: JSON.stringify(payload),
    });
    message.textContent = "Problem updated.";
    message.className = "message success";
    await Promise.all([loadProblems(), loadAdminProblems()]);
    await showProblemDetail(problemID, true);
  } catch (error) {
    message.textContent = error.message;
    message.className = "message error";
  }
}

async function updateProblemStatus(problemID, status, keepDetailOpen = false) {
  try {
    await request(`/api/v1/admin/problems/${problemID}/status`, {
      method: "PATCH",
      headers: authHeaders(),
      body: JSON.stringify({ status }),
    });
    await Promise.all([loadProblems(), loadAdminProblems()]);
    if (keepDetailOpen) {
      await showProblemDetail(problemID, true);
    }
  } catch (error) {
    window.alert(error.message);
  }
}

async function deleteProblem(problemID) {
  if (!window.confirm("Delete this problem permanently?")) {
    return;
  }

  try {
    await request(`/api/v1/admin/problems/${problemID}`, {
      method: "DELETE",
      headers: authHeaders(),
    });
    if (state.selectedProblem?.id === problemID) {
      state.selectedProblem = null;
      nodes.problemDetail.classList.add("hidden");
      nodes.problemDetail.innerHTML = "";
    }
    await Promise.all([loadProblems(), loadAdminProblems()]);
  } catch (error) {
    window.alert(error.message);
  }
}

function createTestCase(isSample = false) {
  return {
    input: "",
    output: "",
    is_sample: isSample,
  };
}

function updateCaseState(mode, index, field, value) {
  const target = mode === "upload" ? state.uploadCases : state.editorCases;
  if (!target[index]) {
    return;
  }
  target[index][field] = value;
}

function removeCase(mode, index) {
  const target = mode === "upload" ? state.uploadCases : state.editorCases;
  if (target.length === 1) {
    return;
  }
  target.splice(index, 1);
  if (mode === "upload") {
    renderUploadCases();
    return;
  }
  renderEditorCases();
}

function initSubmissionEditor() {
  const textarea = document.getElementById("submit-code");
  const highlight = document.getElementById("editor-highlight");
  const gutter = document.getElementById("editor-gutter");
  const autocomplete = document.getElementById("editor-autocomplete");
  const languageSelect = document.getElementById("submit-language");
  const indentSelect = document.getElementById("editor-indent-size");
  const fontSizeSelect = document.getElementById("editor-font-size");
  const wrapToggle = document.getElementById("editor-wrap-lines");
  const formatButton = document.getElementById("editor-format");
  const editor = document.getElementById("code-editor");
  const autocompleteState = {
    items: [],
    selectedIndex: 0,
    range: null,
  };

  if (!textarea || !highlight || !gutter || !autocomplete || !languageSelect || !indentSelect || !fontSizeSelect || !wrapToggle || !formatButton || !editor) {
    return;
  }

  languageSelect.value = state.editorPrefs.language || languageSelect.value;
  textarea.value = state.editorDrafts[languageSelect.value] || defaultCodeForLanguage(languageSelect.value);
  applyEditorPrefs(editor, textarea, highlight, gutter);
  refreshEditorPresentation();

  textarea.addEventListener("input", () => {
    state.editorDrafts[languageSelect.value] = textarea.value;
    persistEditorDrafts();
    refreshEditorPresentation();
    updateAutocomplete(true);
  });

  textarea.addEventListener("scroll", () => {
    highlight.scrollTop = textarea.scrollTop;
    highlight.scrollLeft = textarea.scrollLeft;
    gutter.scrollTop = textarea.scrollTop;
  });

  textarea.addEventListener("click", () => updateAutocomplete(false));
  textarea.addEventListener("keyup", (event) => {
    if (["ArrowUp", "ArrowDown", "Enter", "Tab", "Escape"].includes(event.key)) {
      return;
    }
    updateAutocomplete(false);
  });

  textarea.addEventListener("blur", () => {
    window.setTimeout(hideAutocomplete, 120);
  });

  textarea.addEventListener("keydown", (event) => handleEditorKeydown(event, textarea, Number(indentSelect.value), languageSelect.value));

  languageSelect.addEventListener("change", () => {
    const previousLanguage = state.editorPrefs.language;
    state.editorDrafts[previousLanguage] = textarea.value;
    persistEditorDrafts();
    state.editorPrefs.language = languageSelect.value;
    persistEditorPrefs();
    textarea.value = state.editorDrafts[languageSelect.value] || defaultCodeForLanguage(languageSelect.value);
    state.editorDrafts[languageSelect.value] = textarea.value;
    persistEditorDrafts();
    refreshEditorPresentation();
    updateAutocomplete(true);
  });

  indentSelect.addEventListener("change", () => {
    state.editorPrefs.indentSize = Number(indentSelect.value);
    persistEditorPrefs();
    applyEditorPrefs(editor, textarea, highlight, gutter);
    refreshEditorPresentation();
  });

  fontSizeSelect.addEventListener("change", () => {
    state.editorPrefs.fontSize = Number(fontSizeSelect.value);
    persistEditorPrefs();
    applyEditorPrefs(editor, textarea, highlight, gutter);
  });

  wrapToggle.addEventListener("change", () => {
    state.editorPrefs.wrap = wrapToggle.checked;
    persistEditorPrefs();
    applyEditorPrefs(editor, textarea, highlight, gutter);
    refreshEditorPresentation();
  });

  formatButton.addEventListener("click", () => {
    textarea.value = formatCode(textarea.value, languageSelect.value, Number(indentSelect.value));
    state.editorDrafts[languageSelect.value] = textarea.value;
    persistEditorDrafts();
    refreshEditorPresentation();
    hideAutocomplete();
  });

  function refreshEditorPresentation() {
    const language = languageSelect.value;
    highlight.innerHTML = highlightCode(textarea.value, language);
    gutter.innerHTML = buildLineNumbers(textarea.value);
    highlight.className = `editor-highlight language-${language}`;
    editor.dataset.language = language;
  }

  function updateAutocomplete(resetSelection) {
    const suggestionSet = getAutocompleteEntries(languageSelect.value);
    const context = getAutocompleteContext(textarea.value, textarea.selectionStart);
    if (!context || (!context.prefix && !resetSelection)) {
      hideAutocomplete();
      return;
    }

    const filtered = suggestionSet.filter((item) => item.label.toLowerCase().startsWith(context.prefix.toLowerCase()))
      .slice(0, 8);
    if (!filtered.length) {
      hideAutocomplete();
      return;
    }

    autocompleteState.items = filtered;
    autocompleteState.range = context;
    autocompleteState.selectedIndex = resetSelection ? 0 : Math.min(autocompleteState.selectedIndex, filtered.length - 1);
    renderAutocomplete();
  }

  function renderAutocomplete() {
    autocomplete.innerHTML = autocompleteState.items.map((item, index) => `
      <button class="autocomplete-item ${index === autocompleteState.selectedIndex ? "is-active" : ""}" data-autocomplete-index="${index}" type="button">
        <strong>${escapeHtml(item.label)}</strong>
        <span>${escapeHtml(item.detail || "")}</span>
      </button>
    `).join("");
    autocomplete.classList.remove("hidden");

    [...autocomplete.querySelectorAll("[data-autocomplete-index]")].forEach((button) => {
      button.addEventListener("mousedown", (event) => {
        event.preventDefault();
        autocompleteState.selectedIndex = Number(button.dataset.autocompleteIndex);
        applyAutocomplete();
      });
    });
  }

  function hideAutocomplete() {
    autocompleteState.items = [];
    autocompleteState.range = null;
    autocomplete.classList.add("hidden");
    autocomplete.innerHTML = "";
  }

  function applyAutocomplete() {
    const current = autocompleteState.items[autocompleteState.selectedIndex];
    if (!current || !autocompleteState.range) {
      return;
    }

    const before = textarea.value.slice(0, autocompleteState.range.start);
    const after = textarea.value.slice(autocompleteState.range.end);
    textarea.value = before + current.insert + after;
    const cursor = before.length + current.cursorOffset;
    textarea.selectionStart = cursor;
    textarea.selectionEnd = cursor;
    state.editorDrafts[languageSelect.value] = textarea.value;
    persistEditorDrafts();
    refreshEditorPresentation();
    hideAutocomplete();
    textarea.focus();
  }

  function handleEditorKeydown(event, editorInput, indentSize, language) {
    if (!event.ctrlKey && !event.metaKey && autocompleteState.items.length) {
      if (event.key === "ArrowDown") {
        event.preventDefault();
        autocompleteState.selectedIndex = (autocompleteState.selectedIndex + 1) % autocompleteState.items.length;
        renderAutocomplete();
        return;
      }
      if (event.key === "ArrowUp") {
        event.preventDefault();
        autocompleteState.selectedIndex = (autocompleteState.selectedIndex - 1 + autocompleteState.items.length) % autocompleteState.items.length;
        renderAutocomplete();
        return;
      }
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        applyAutocomplete();
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        hideAutocomplete();
        return;
      }
    }

    if (event.ctrlKey && event.key === " ") {
      event.preventDefault();
      updateAutocomplete(true);
      return;
    }

    if (handlePairInsertion(event, editorInput)) {
      state.editorDrafts[language] = editorInput.value;
      persistEditorDrafts();
      refreshEditorPresentation();
      hideAutocomplete();
      return;
    }

    if (event.key === "Tab") {
      event.preventDefault();
      if (event.shiftKey) {
        outdentSelection(editorInput, indentSize);
        hideAutocomplete();
        return;
      }
      indentSelection(editorInput, indentSize);
      hideAutocomplete();
      return;
    }

    if (event.key === "Enter") {
      event.preventDefault();
      insertNewlineWithIndent(editorInput, indentSize);
      hideAutocomplete();
    }
  }
}

function applyEditorPrefs(editor, textarea, highlight, gutter) {
  const fontSize = `${state.editorPrefs.fontSize}px`;
  const tabSize = String(state.editorPrefs.indentSize);

  editor.classList.toggle("is-wrap", state.editorPrefs.wrap);
  textarea.style.fontSize = fontSize;
  highlight.style.fontSize = fontSize;
  gutter.style.fontSize = fontSize;
  textarea.style.tabSize = tabSize;
  highlight.style.tabSize = tabSize;
  textarea.wrap = state.editorPrefs.wrap ? "soft" : "off";
}

function buildLineNumbers(code) {
  const lines = Math.max(1, code.split("\n").length);
  return Array.from({ length: lines }, (_, index) => `<span>${index + 1}</span>`).join("");
}

function indentSelection(textarea, indentSize) {
  const indent = " ".repeat(indentSize);
  const { selectionStart, selectionEnd, value } = textarea;
  const lineStart = value.lastIndexOf("\n", selectionStart - 1) + 1;
  const selected = value.slice(lineStart, selectionEnd);
  const lines = selected.split("\n");
  const replaced = lines.map((line) => indent + line).join("\n");

  textarea.value = value.slice(0, lineStart) + replaced + value.slice(selectionEnd);
  textarea.selectionStart = selectionStart + indentSize;
  textarea.selectionEnd = selectionEnd + indentSize * lines.length;
  textarea.dispatchEvent(new Event("input"));
}

function outdentSelection(textarea, indentSize) {
  const { selectionStart, selectionEnd, value } = textarea;
  const lineStart = value.lastIndexOf("\n", selectionStart - 1) + 1;
  const selected = value.slice(lineStart, selectionEnd);
  const lines = selected.split("\n");

  let removed = 0;
  const replaced = lines.map((line, index) => {
    const current = Math.min(indentSize, line.match(/^ */)?.[0].length || 0);
    if (current > 0) {
      removed += current;
      return line.slice(current);
    }
    return line;
  }).join("\n");

  const firstLineRemoved = Math.min(indentSize, value.slice(lineStart).match(/^ */)?.[0].length || 0);
  textarea.value = value.slice(0, lineStart) + replaced + value.slice(selectionEnd);
  textarea.selectionStart = Math.max(lineStart, selectionStart - firstLineRemoved);
  textarea.selectionEnd = Math.max(textarea.selectionStart, selectionEnd - removed);
  textarea.dispatchEvent(new Event("input"));
}

function insertNewlineWithIndent(textarea, indentSize) {
  const { selectionStart, selectionEnd, value } = textarea;
  const beforeChar = value[selectionStart - 1] || "";
  const afterChar = value[selectionStart] || "";
  const lineStart = value.lastIndexOf("\n", selectionStart - 1) + 1;
  const currentLine = value.slice(lineStart, selectionStart);
  const baseIndent = currentLine.match(/^\s*/)?.[0] || "";
  const trimmed = currentLine.trimEnd();
  const extraIndent = shouldIncreaseIndent(trimmed) ? " ".repeat(indentSize) : "";
  if (selectionStart === selectionEnd && isBracketPair(beforeChar, afterChar)) {
    const insertText = `\n${baseIndent}${extraIndent}\n${baseIndent}`;
    textarea.value = value.slice(0, selectionStart) + insertText + value.slice(selectionEnd);
    const nextCursor = selectionStart + 1 + baseIndent.length + extraIndent.length;
    textarea.selectionStart = nextCursor;
    textarea.selectionEnd = nextCursor;
    textarea.dispatchEvent(new Event("input"));
    return;
  }

  const insertText = `\n${baseIndent}${extraIndent}`;

  textarea.value = value.slice(0, selectionStart) + insertText + value.slice(selectionEnd);
  const nextCursor = selectionStart + insertText.length;
  textarea.selectionStart = nextCursor;
  textarea.selectionEnd = nextCursor;
  textarea.dispatchEvent(new Event("input"));
}

function shouldIncreaseIndent(trimmedLine) {
  return /[\{\[\(]$/.test(trimmedLine) || /:\s*$/.test(trimmedLine);
}

function formatCode(code, language, indentSize) {
  switch (language) {
  case "python":
    return formatPythonCode(code, indentSize);
  case "java":
  case "cpp":
  default:
    return formatBracedCode(code, indentSize);
  }
}

function defaultCodeForLanguage(language) {
  switch (language) {
  case "java":
    return [
      "import java.util.*;",
      "",
      "public class Main {",
      "    public static void main(String[] args) {",
      "        Scanner scanner = new Scanner(System.in);",
      "    }",
      "}",
    ].join("\n");
  case "python":
    return [
      "def solve():",
      "    pass",
      "",
      "",
      "if __name__ == \"__main__\":",
      "    solve()",
    ].join("\n");
  default:
    return [
      "#include <bits/stdc++.h>",
      "using namespace std;",
      "",
      "int main() {",
      "    ios::sync_with_stdio(false);",
      "    cin.tie(nullptr);",
      "",
      "    return 0;",
      "}",
    ].join("\n");
  }
}

function formatBracedCode(code, indentSize) {
  const indent = " ".repeat(indentSize);
  const lines = normalizeEditorText(code).split("\n");
  let depth = 0;

  return trimTrailingBlankLines(lines.map((line) => {
    const raw = line.replaceAll("\t", indent).trim();
    if (!raw) {
      return "";
    }

    const dedentFirst = /^[\}\]\)]/.test(raw);
    if (dedentFirst) {
      depth = Math.max(0, depth - 1);
    }

    const formatted = indent.repeat(depth) + raw;
    depth += countStructuralDelta(raw) + (dedentFirst ? 1 : 0);
    if (dedentFirst) {
      depth = Math.max(depth, 0);
    }
    return formatted;
  })).join("\n");
}

function formatPythonCode(code, indentSize) {
  const indent = " ".repeat(indentSize);
  const lines = normalizeEditorText(code).split("\n");
  let depth = 0;

  return trimTrailingBlankLines(lines.map((line) => {
    const raw = line.replaceAll("\t", indent).trim();
    if (!raw) {
      return "";
    }

    if (/^(elif|else|except|finally)\b/.test(raw)) {
      depth = Math.max(0, depth - 1);
    }

    const formatted = indent.repeat(depth) + raw;
    if (/:$/.test(raw) && !raw.startsWith("#")) {
      depth += 1;
    }
    return formatted;
  })).join("\n");
}

function normalizeEditorText(code) {
  return code.replaceAll("\r\n", "\n").replaceAll("\r", "\n").replace(/[ \t]+$/gm, "");
}

function trimTrailingBlankLines(code) {
  return code.replace(/\n{3,}$/g, "\n\n").replace(/\s+$/g, "");
}

function countStructuralDelta(line) {
  const opens = (line.match(/[\{\[\(]/g) || []).length;
  const closes = (line.match(/[\}\]\)]/g) || []).length;
  return opens - closes;
}

function getAutocompleteEntries(language) {
  const common = [
    { label: "if", insert: "if () {\n    \n}", cursorOffset: 4, detail: "Conditional block" },
    { label: "for", insert: "for () {\n    \n}", cursorOffset: 5, detail: "Loop block" },
    { label: "while", insert: "while () {\n    \n}", cursorOffset: 7, detail: "Loop block" },
  ];

  if (language === "java") {
    return [
      { label: "public class Main", insert: "public class Main {\n    public static void main(String[] args) {\n        \n    }\n}", cursorOffset: 67, detail: "Main entry" },
      { label: "Scanner", insert: "Scanner scanner = new Scanner(System.in);", cursorOffset: 40, detail: "Input reader" },
      { label: "System.out.println", insert: "System.out.println();", cursorOffset: 19, detail: "Print line" },
      ...common,
      ...[...getLanguageKeywords(language)].map((keyword) => ({ label: keyword, insert: keyword, cursorOffset: keyword.length, detail: "Keyword" })),
    ];
  }

  if (language === "python") {
    return [
      { label: "def solve", insert: "def solve():\n    \n", cursorOffset: 16, detail: "Function scaffold" },
      { label: "if __name__", insert: "if __name__ == \"__main__\":\n    solve()", cursorOffset: 37, detail: "Entry point" },
      { label: "print", insert: "print()", cursorOffset: 6, detail: "Print call" },
      ...common.map((item) => ({
        ...item,
        insert: item.label === "if" ? "if :\n    " : item.label === "for" ? "for  in :\n    " : "while :\n    ",
        cursorOffset: item.label === "if" ? 3 : item.label === "for" ? 5 : 6,
      })),
      ...[...getLanguageKeywords(language)].map((keyword) => ({ label: keyword, insert: keyword, cursorOffset: keyword.length, detail: "Keyword" })),
    ];
  }

  return [
    { label: "#include <bits/stdc++.h>", insert: "#include <bits/stdc++.h>", cursorOffset: 24, detail: "Header include" },
    { label: "vector<int>", insert: "vector<int>", cursorOffset: 11, detail: "Container" },
    { label: "cout <<", insert: "cout << ;", cursorOffset: 8, detail: "Output statement" },
    ...common,
    ...[...getLanguageKeywords(language)].map((keyword) => ({ label: keyword, insert: keyword, cursorOffset: keyword.length, detail: "Keyword" })),
  ];
}

function getAutocompleteContext(code, cursor) {
  const left = code.slice(0, cursor);
  const match = left.match(/([A-Za-z_][A-Za-z0-9_]*)$/);
  if (!match) {
    return null;
  }
  return {
    prefix: match[1],
    start: cursor - match[1].length,
    end: cursor,
  };
}

function handlePairInsertion(event, textarea) {
  const openToClose = {
    "(": ")",
    "[": "]",
    "{": "}",
    "\"": "\"",
    "'": "'",
  };
  const closeToOpen = {
    ")": "(",
    "]": "[",
    "}": "{",
    "\"": "\"",
    "'": "'",
  };
  const { selectionStart, selectionEnd, value } = textarea;

  if (Object.hasOwn(closeToOpen, event.key) && selectionStart === selectionEnd && value[selectionStart] === event.key) {
    event.preventDefault();
    textarea.selectionStart = selectionStart + 1;
    textarea.selectionEnd = selectionStart + 1;
    return true;
  }

  if (Object.hasOwn(openToClose, event.key) && !event.ctrlKey && !event.metaKey && !event.altKey) {
    event.preventDefault();
    const closing = openToClose[event.key];
    const selected = value.slice(selectionStart, selectionEnd);
    const replacement = `${event.key}${selected}${closing}`;
    textarea.value = value.slice(0, selectionStart) + replacement + value.slice(selectionEnd);
    const cursor = selected ? selectionEnd + 2 : selectionStart + 1;
    textarea.selectionStart = cursor;
    textarea.selectionEnd = cursor;
    textarea.dispatchEvent(new Event("input"));
    return true;
  }

  if (event.key === "Backspace" && selectionStart === selectionEnd) {
    const previous = value[selectionStart - 1] || "";
    const next = value[selectionStart] || "";
    if (isBracketPair(previous, next)) {
      event.preventDefault();
      textarea.value = value.slice(0, selectionStart - 1) + value.slice(selectionStart + 1);
      textarea.selectionStart = selectionStart - 1;
      textarea.selectionEnd = selectionStart - 1;
      textarea.dispatchEvent(new Event("input"));
      return true;
    }
  }

  return false;
}

function isBracketPair(left, right) {
  return (left === "(" && right === ")")
    || (left === "[" && right === "]")
    || (left === "{" && right === "}")
    || (left === "\"" && right === "\"")
    || (left === "'" && right === "'");
}

function highlightCode(code, language) {
  const tokens = tokenizeCode(code, language);
  return tokens.map((token) => {
    const content = escapeHtml(token.value);
    if (token.type === "plain") {
      return content || " ";
    }
    return `<span class="token token-${token.type}">${content || " "}</span>`;
  }).join("");
}

function tokenizeCode(code, language) {
  const keywords = getLanguageKeywords(language);
  const tokens = [];
  let index = 0;

  while (index < code.length) {
    const current = code[index];
    const next = code[index + 1] || "";

    if (language !== "python" && current === "/" && next === "/") {
      const end = code.indexOf("\n", index);
      const stop = end === -1 ? code.length : end;
      tokens.push({ type: "comment", value: code.slice(index, stop) });
      index = stop;
      continue;
    }

    if (language !== "python" && current === "/" && next === "*") {
      const end = code.indexOf("*/", index + 2);
      const stop = end === -1 ? code.length : end + 2;
      tokens.push({ type: "comment", value: code.slice(index, stop) });
      index = stop;
      continue;
    }

    if (language === "python" && current === "#") {
      const end = code.indexOf("\n", index);
      const stop = end === -1 ? code.length : end;
      tokens.push({ type: "comment", value: code.slice(index, stop) });
      index = stop;
      continue;
    }

    if (current === "\"" || current === "'" || current === "`") {
      const stop = readStringToken(code, index, current);
      tokens.push({ type: "string", value: code.slice(index, stop) });
      index = stop;
      continue;
    }

    if (/\d/.test(current)) {
      const stop = readNumberToken(code, index);
      tokens.push({ type: "number", value: code.slice(index, stop) });
      index = stop;
      continue;
    }

    if (/[A-Za-z_]/.test(current)) {
      const stop = readIdentifierToken(code, index);
      const value = code.slice(index, stop);
      tokens.push({ type: keywords.has(value) ? "keyword" : "plain", value });
      index = stop;
      continue;
    }

    if (/[\{\}\[\]\(\)=<>!+\-*%&|:;,.]/.test(current)) {
      tokens.push({ type: "operator", value: current });
      index += 1;
      continue;
    }

    tokens.push({ type: "plain", value: current });
    index += 1;
  }

  return tokens;
}

function readStringToken(code, start, quote) {
  let index = start + 1;
  while (index < code.length) {
    if (code[index] === "\\") {
      index += 2;
      continue;
    }
    if (code[index] === quote) {
      return index + 1;
    }
    index += 1;
  }
  return code.length;
}

function readNumberToken(code, start) {
  let index = start;
  while (index < code.length && /[\d._]/.test(code[index])) {
    index += 1;
  }
  return index;
}

function readIdentifierToken(code, start) {
  let index = start;
  while (index < code.length && /[\w]/.test(code[index])) {
    index += 1;
  }
  return index;
}

function getLanguageKeywords(language) {
  if (language === "java") {
    return new Set(["public", "private", "protected", "class", "static", "void", "int", "long", "double", "float", "boolean", "char", "new", "return", "if", "else", "for", "while", "break", "continue", "import", "package", "try", "catch", "throws", "throw", "Scanner"]);
  }
  if (language === "python") {
    return new Set(["def", "return", "if", "elif", "else", "for", "while", "break", "continue", "import", "from", "as", "class", "try", "except", "finally", "with", "lambda", "pass", "in", "is", "None", "True", "False"]);
  }
  return new Set(["int", "long", "double", "float", "bool", "char", "void", "string", "vector", "map", "set", "queue", "stack", "pair", "auto", "using", "namespace", "return", "if", "else", "for", "while", "break", "continue", "include", "class", "struct", "public", "private", "protected", "template", "typename", "const", "nullptr", "switch", "case"]);
}

function loadEditorPrefs() {
  try {
    const saved = JSON.parse(localStorage.getItem("algojudge_editor_prefs") || "{}");
    return {
      language: saved.language || "cpp",
      fontSize: [13, 14, 16].includes(saved.fontSize) ? saved.fontSize : 14,
      indentSize: [2, 4].includes(saved.indentSize) ? saved.indentSize : 4,
      wrap: Boolean(saved.wrap),
    };
  } catch {
    return { language: "cpp", fontSize: 14, indentSize: 4, wrap: false };
  }
}

function persistEditorPrefs() {
  localStorage.setItem("algojudge_editor_prefs", JSON.stringify(state.editorPrefs));
}

function loadEditorDrafts() {
  try {
    const drafts = JSON.parse(localStorage.getItem("algojudge_editor_drafts") || "{}");
    return typeof drafts === "object" && drafts ? drafts : {};
  } catch {
    return {};
  }
}

function persistEditorDrafts() {
  localStorage.setItem("algojudge_editor_drafts", JSON.stringify(state.editorDrafts));
}

function normalizeProblem(problem) {
  const normalized = { ...problem };
  normalized.limits = normalized.limits || {
    time: normalized.time_limit,
    memory: normalized.memory_limit,
  };
  if (!normalized.samples && normalized.test_cases) {
    normalized.samples = normalized.test_cases.filter((item) => item.is_sample);
  }
  return normalized;
}

async function trackSubmission(submissionID) {
  const message = document.getElementById("submit-message");
  const resultBox = document.getElementById("submission-result");

  for (let attempt = 0; attempt < 20; attempt += 1) {
    await delay(900);

    try {
      const submission = await request(`/api/v1/submissions/${submissionID}`);
      resultBox.classList.remove("hidden");
      renderSubmissionResult(submission);
      loadStatus();
      loadRecentSubmissions();

      if (!isSubmissionActive(submission.status)) {
        message.textContent = `Submission finished with ${submission.status}.`;
        message.className = submission.status === "AC" ? "message success" : "message error";
        return;
      }
    } catch (error) {
      message.textContent = error.message;
      message.className = "message error";
      return;
    }
  }

  message.textContent = "Submission is still running. Check the Status page for updates.";
  message.className = "message";
}

function renderSubmissionResult(submission) {
  const resultBox = document.getElementById("submission-result");
  resultBox.innerHTML = `
    <strong>Status: ${escapeHtml(submission.status)}</strong>
    <p>Runtime: ${submission.runtime} ms</p>
    <p>Memory: ${submission.memory} KB</p>
    <p>Submission ID: ${submission.id}</p>
    ${submission.error_msg ? `<p>Error: ${escapeHtml(submission.error_msg)}</p>` : ""}
  `;
}

function isSubmissionActive(status) {
  return status === "Pending" || status === "Judging";
}

function delay(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function statusLabel(status) {
  switch (Number(status)) {
  case 1:
    return "Published";
  case 2:
    return "Hidden";
  default:
    return "Pending";
  }
}

function statusClassName(status) {
  switch (Number(status)) {
  case 1:
    return "published";
  case 2:
    return "hidden";
  default:
    return "pending";
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
  state.adminProblems = [];
  state.submissions = [];
  localStorage.removeItem("algojudge_token");
  localStorage.removeItem("algojudge_user");
  if (updateUI) {
    refreshAuthUI();
    renderRecentSubmissions();
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

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleString();
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

(function () {
  const storageKeys = {
    token: "infohub.token",
    userID: "infohub.user_id",
  };

  const state = {
    token: "",
    selectedReportName: "",
  };

  const elements = {
    tokenInput: document.getElementById("tokenInput"),
    runReportButton: document.getElementById("runReportButton"),
    refreshButton: document.getElementById("refreshButton"),
    statusMessage: document.getElementById("statusMessage"),
    generatedAtLabel: document.getElementById("generatedAtLabel"),
    itemCountValue: document.getElementById("itemCountValue"),
    displayCountValue: document.getElementById("displayCountValue"),
    priorityCountValue: document.getElementById("priorityCountValue"),
    topTitlesList: document.getElementById("topTitlesList"),
    decisionList: document.getElementById("decisionList"),
    reportMarkdown: document.getElementById("reportMarkdown"),
    historyList: document.getElementById("historyList"),
    userIdInput: document.getElementById("userIdInput"),
    tagsInput: document.getElementById("tagsInput"),
    sourcesInput: document.getElementById("sourcesInput"),
    keywordsInput: document.getElementById("keywordsInput"),
    weightTagInput: document.getElementById("weightTagInput"),
    weightSourceInput: document.getElementById("weightSourceInput"),
    weightKeywordInput: document.getElementById("weightKeywordInput"),
    loadPreferenceButton: document.getElementById("loadPreferenceButton"),
    savePreferenceButton: document.getElementById("savePreferenceButton"),
  };

  function init() {
    state.token = window.localStorage.getItem(storageKeys.token) || "";
    elements.tokenInput.value = state.token;
    elements.userIdInput.value = window.localStorage.getItem(storageKeys.userID) || "";

    elements.tokenInput.addEventListener("input", () => {
      state.token = elements.tokenInput.value.trim();
      window.localStorage.setItem(storageKeys.token, state.token);
    });

    elements.userIdInput.addEventListener("input", () => {
      window.localStorage.setItem(storageKeys.userID, elements.userIdInput.value.trim());
    });

    elements.runReportButton.addEventListener("click", runReport);
    elements.refreshButton.addEventListener("click", refreshAll);
    elements.loadPreferenceButton.addEventListener("click", loadPreference);
    elements.savePreferenceButton.addEventListener("click", savePreference);

    refreshAll();
  }

  async function apiRequest(path, options) {
    const request = Object.assign({ headers: {} }, options || {});
    request.headers = Object.assign({}, request.headers);
    if (state.token) {
      request.headers.Authorization = `Bearer ${state.token}`;
    }
    if (request.body && !request.headers["Content-Type"]) {
      request.headers["Content-Type"] = "application/json";
    }

    const response = await fetch(path, request);
    let payload = null;
    const text = await response.text();
    if (text) {
      try {
        payload = JSON.parse(text);
      } catch (error) {
        payload = { raw: text };
      }
    }

    if (!response.ok) {
      const message = buildErrorMessage(response.status, payload);
      const requestError = new Error(message);
      requestError.status = response.status;
      requestError.payload = payload;
      throw requestError;
    }

    return payload;
  }

  function buildErrorMessage(status, payload) {
    const serverMessage = payload && payload.error ? payload.error : "";
    if (status === 401) {
      return "鉴权失败，请检查访问 Token。";
    }
    if (status === 404) {
      return serverMessage || "未找到对应数据。";
    }
    if (status >= 500) {
      return serverMessage || "服务端处理失败，请稍后重试。";
    }
    return serverMessage || "请求失败，请检查输入后重试。";
  }

  function setStatus(message, isError) {
    elements.statusMessage.textContent = message;
    elements.statusMessage.classList.toggle("is-error", Boolean(isError));
  }

  function collectPreference() {
    return {
      tags: splitList(elements.tagsInput.value),
      sources: splitList(elements.sourcesInput.value),
      keywords: splitList(elements.keywordsInput.value),
      weights: {
        tag: parseWeight(elements.weightTagInput.value),
        source: parseWeight(elements.weightSourceInput.value),
        keyword: parseWeight(elements.weightKeywordInput.value),
      },
    };
  }

  function splitList(value) {
    return value
      .split(/[\n,，]/)
      .map((item) => item.trim())
      .filter(Boolean);
  }

  function parseWeight(value) {
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0;
  }

  async function refreshAll() {
    setStatus("正在刷新数据...", false);
    await Promise.all([loadLatestReport(), loadReportHistory()]);
    setStatus("数据已刷新。", false);
  }

  async function runReport() {
    setStatus("正在生成日报...", false);
    const userID = elements.userIdInput.value.trim();
    const body = {};
    if (userID) {
      body.user_id = userID;
    }
    body.preference = collectPreference();

    try {
      await apiRequest("/reports/run", {
        method: "POST",
        body: JSON.stringify(body),
      });
      await refreshAll();
      setStatus("日报已生成，界面已更新。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function loadLatestReport() {
    try {
      const report = await apiRequest("/reports/latest");
      state.selectedReportName = "";
      renderReport(report);
    } catch (error) {
      if (error.status === 404) {
        renderEmptyReport();
        return;
      }
      setStatus(error.message, true);
      throw error;
    }
  }

  async function loadReportHistory() {
    try {
      const payload = await apiRequest("/reports");
      renderHistory(payload && payload.reports ? payload.reports : []);
    } catch (error) {
      if (error.status === 404) {
        renderHistory([]);
        return;
      }
      setStatus(error.message, true);
      throw error;
    }
  }

  async function openHistoryReport(name) {
    setStatus("正在加载历史日报...", false);
    try {
      const report = await apiRequest(`/reports/${encodeURIComponent(name)}`);
      state.selectedReportName = name;
      renderReport(report);
      highlightHistory(name);
      setStatus(`已加载历史日报 ${name}。`, false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function loadPreference() {
    const userID = elements.userIdInput.value.trim();
    if (!userID) {
      setStatus("请先输入用户 ID。", true);
      return;
    }

    setStatus("正在读取偏好...", false);
    try {
      const payload = await apiRequest(`/preferences/${encodeURIComponent(userID)}`);
      fillPreference(payload);
      setStatus("已读取已保存偏好。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function savePreference() {
    const userID = elements.userIdInput.value.trim();
    if (!userID) {
      setStatus("请先输入用户 ID。", true);
      return;
    }

    setStatus("正在保存偏好...", false);
    try {
      await apiRequest(`/preferences/${encodeURIComponent(userID)}`, {
        method: "PUT",
        body: JSON.stringify(collectPreference()),
      });
      setStatus("偏好已保存。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  function fillPreference(payload) {
    elements.tagsInput.value = (payload.tags || []).join(", ");
    elements.sourcesInput.value = (payload.sources || []).join(", ");
    elements.keywordsInput.value = (payload.keywords || []).join(", ");
    elements.weightTagInput.value = payload.weights ? payload.weights.tag : 0;
    elements.weightSourceInput.value = payload.weights ? payload.weights.source : 0;
    elements.weightKeywordInput.value = payload.weights ? payload.weights.keyword : 0;
  }

  function renderEmptyReport() {
    elements.generatedAtLabel.textContent = "暂无日报";
    elements.itemCountValue.textContent = "-";
    elements.displayCountValue.textContent = "-";
    elements.priorityCountValue.textContent = "-";
    elements.topTitlesList.innerHTML = "<li>暂无数据</li>";
    elements.decisionList.textContent = "暂无决策摘要";
    elements.decisionList.className = "decision-list empty-state";
    elements.reportMarkdown.textContent = "暂无日报，请先生成或选择历史记录。";
    setStatus("当前还没有日报，先点击“生成日报”。", false);
  }

  function renderReport(report) {
    elements.generatedAtLabel.textContent = formatDateTime(report.generated_at);
    elements.itemCountValue.textContent = String((report.items || []).length);
    elements.displayCountValue.textContent = String(report.display_count || 0);
    elements.priorityCountValue.textContent = String(report.high_priority_count || 0);
    renderTopTitles(report.top_priority_items || []);
    renderDecisions(report.decision_summary || []);
    elements.reportMarkdown.textContent = report.markdown || "暂无日报内容";
  }

  function renderTopTitles(titles) {
    if (!titles.length) {
      elements.topTitlesList.innerHTML = "<li>暂无重点标题</li>";
      return;
    }

    elements.topTitlesList.innerHTML = titles
      .map((title) => `<li>${escapeHTML(title)}</li>`)
      .join("");
  }

  function renderDecisions(items) {
    if (!items.length) {
      elements.decisionList.textContent = "暂无决策摘要";
      elements.decisionList.className = "decision-list empty-state";
      return;
    }

    elements.decisionList.className = "decision-list";
    elements.decisionList.innerHTML = items
      .map((item) => {
        const tags = Array.isArray(item.tags) && item.tags.length
          ? `<div class="tag-list">${item.tags.map((tag) => `<span class="tag">${escapeHTML(tag)}</span>`).join("")}</div>`
          : "";
        return `
          <article class="decision-item">
            <h3>${escapeHTML(item.title || "未命名条目")}</h3>
            <div class="decision-meta">
              <span>来源：${escapeHTML(item.source || "未知来源")}</span>
              <span>评分：${escapeHTML(String(item.score || 0))}</span>
              <span>建议：${escapeHTML(item.action || "待判断")}</span>
            </div>
            <p class="decision-summary">${escapeHTML(item.summary || "暂无摘要")}</p>
            ${tags}
          </article>
        `;
      })
      .join("");
  }

  function renderHistory(reports) {
    if (!reports.length) {
      elements.historyList.textContent = "暂无历史日报";
      elements.historyList.className = "history-list empty-state";
      return;
    }

    elements.historyList.className = "history-list";
    elements.historyList.innerHTML = reports
      .map((report) => `
        <button class="history-item${state.selectedReportName === report.name ? " is-active" : ""}" data-report-name="${escapeHTML(report.name)}" type="button">
          <h3>${escapeHTML(formatName(report.name))}</h3>
          <div class="history-meta">
            <span>采集 ${escapeHTML(String(report.item_count || 0))}</span>
            <span>展示 ${escapeHTML(String(report.display_count || 0))}</span>
            <span>高优先级 ${escapeHTML(String(report.high_priority_count || 0))}</span>
          </div>
          <div class="tag-list">
            ${(report.top_titles || []).map((title) => `<span class="tag">${escapeHTML(title)}</span>`).join("")}
          </div>
        </button>
      `)
      .join("");

    Array.from(elements.historyList.querySelectorAll("[data-report-name]")).forEach((button) => {
      button.addEventListener("click", () => openHistoryReport(button.getAttribute("data-report-name")));
    });
  }

  function highlightHistory(name) {
    Array.from(elements.historyList.querySelectorAll("[data-report-name]")).forEach((button) => {
      button.classList.toggle("is-active", button.getAttribute("data-report-name") === name);
    });
  }

  function formatName(name) {
    return name || "未命名日报";
  }

  function formatDateTime(value) {
    if (!value) {
      return "暂无时间";
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return new Intl.DateTimeFormat("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }).format(date);
  }

  function escapeHTML(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  init();
})();

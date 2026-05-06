(function () {
  const storageKeys = {
    token: "infohub.token",
    userID: "infohub.user_id",
  };

  const state = {
    token: "",
    selectedReportName: "",
    selectedSearchName: "",
  };

  const elements = {
    tokenInput: document.getElementById("tokenInput"),
    refreshButton: document.getElementById("refreshButton"),
    statusMessage: document.getElementById("statusMessage"),
    runReportButton: document.getElementById("runReportButton"),
    generatedAtLabel: document.getElementById("generatedAtLabel"),
    itemCountValue: document.getElementById("itemCountValue"),
    displayCountValue: document.getElementById("displayCountValue"),
    priorityCountValue: document.getElementById("priorityCountValue"),
    topTitlesList: document.getElementById("topTitlesList"),
    decisionList: document.getElementById("decisionList"),
    reportMarkdown: document.getElementById("reportMarkdown"),
    copyMarkdownButton: document.getElementById("copyMarkdownButton"),
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
    searchQueryInput: document.getElementById("searchQueryInput"),
    runSearchButton: document.getElementById("runSearchButton"),
    copySearchMarkdownButton: document.getElementById("copySearchMarkdownButton"),
    searchHistoryList: document.getElementById("searchHistoryList"),
    searchGeneratedAtLabel: document.getElementById("searchGeneratedAtLabel"),
    searchQueryLabel: document.getElementById("searchQueryLabel"),
    searchItemCountValue: document.getElementById("searchItemCountValue"),
    searchDisplayCountValue: document.getElementById("searchDisplayCountValue"),
    searchPriorityCountValue: document.getElementById("searchPriorityCountValue"),
    searchDecisionList: document.getElementById("searchDecisionList"),
    searchMarkdown: document.getElementById("searchMarkdown"),
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

    elements.refreshButton.addEventListener("click", refreshAll);
    elements.runReportButton.addEventListener("click", runReport);
    elements.runSearchButton.addEventListener("click", runSearch);
    elements.copyMarkdownButton.addEventListener("click", () => copyText(elements.reportMarkdown.textContent || "", elements.copyMarkdownButton));
    elements.copySearchMarkdownButton.addEventListener("click", () => copyText(elements.searchMarkdown.textContent || "", elements.copySearchMarkdownButton));
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
    const text = await response.text();
    let payload = null;
    if (text) {
      try {
        payload = JSON.parse(text);
      } catch (error) {
        payload = { raw: text };
      }
    }

    if (!response.ok) {
      const error = new Error((payload && payload.error) || `请求失败: ${response.status}`);
      error.status = response.status;
      throw error;
    }

    return payload;
  }

  function setStatus(message, isError) {
    elements.statusMessage.textContent = message;
    elements.statusMessage.classList.toggle("is-error", Boolean(isError));
  }

  async function refreshAll() {
    setStatus("正在刷新数据...", false);
    try {
      await Promise.all([loadLatestReport(), loadReportHistory(), loadLatestSearch(), loadSearchHistory()]);
      setStatus("数据已刷新。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function runReport() {
    setStatus("正在生成日报...", false);
    const body = { preference: collectPreference() };
    const userID = elements.userIdInput.value.trim();
    if (userID) {
      body.user_id = userID;
    }

    try {
      await apiRequest("/reports/run", { method: "POST", body: JSON.stringify(body) });
      await Promise.all([loadLatestReport(), loadReportHistory()]);
      setStatus("日报已生成。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function runSearch() {
    const query = elements.searchQueryInput.value.trim();
    if (!query) {
      setStatus("请输入搜索关键词。", true);
      return;
    }

    setStatus(`正在搜索“${query}”...`, false);
    const body = { query, preference: collectPreference() };
    const userID = elements.userIdInput.value.trim();
    if (userID) {
      body.user_id = userID;
    }

    try {
      await apiRequest("/search", { method: "POST", body: JSON.stringify(body) });
      await Promise.all([loadLatestSearch(), loadSearchHistory()]);
      setStatus(`关键词“${query}”搜索完成。`, false);
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
      throw error;
    }
  }

  async function loadReportHistory() {
    try {
      const payload = await apiRequest("/reports");
      renderReportHistory(payload && payload.reports ? payload.reports : []);
    } catch (error) {
      if (error.status === 404) {
        renderReportHistory([]);
        return;
      }
      throw error;
    }
  }

  async function openReport(name) {
    try {
      const report = await apiRequest(`/reports/${encodeURIComponent(name)}`);
      state.selectedReportName = name;
      renderReport(report);
      highlightHistory(elements.historyList, name);
      setStatus(`已加载日报 ${name}。`, false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  async function loadLatestSearch() {
    try {
      const result = await apiRequest("/searches/latest");
      state.selectedSearchName = "";
      renderSearch(result);
    } catch (error) {
      if (error.status === 404) {
        renderEmptySearch();
        return;
      }
      throw error;
    }
  }

  async function loadSearchHistory() {
    try {
      const payload = await apiRequest("/searches");
      renderSearchHistory(payload && payload.searches ? payload.searches : []);
    } catch (error) {
      if (error.status === 404) {
        renderSearchHistory([]);
        return;
      }
      throw error;
    }
  }

  async function openSearch(name) {
    try {
      const result = await apiRequest(`/searches/${encodeURIComponent(name)}`);
      state.selectedSearchName = name;
      renderSearch(result);
      highlightHistory(elements.searchHistoryList, name);
      setStatus(`已加载搜索记录 ${name}。`, false);
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

    try {
      const payload = await apiRequest(`/preferences/${encodeURIComponent(userID)}`);
      fillPreference(payload);
      setStatus("已读取偏好设置。", false);
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

    try {
      await apiRequest(`/preferences/${encodeURIComponent(userID)}`, {
        method: "PUT",
        body: JSON.stringify(collectPreference()),
      });
      setStatus("偏好设置已保存。", false);
    } catch (error) {
      setStatus(error.message, true);
    }
  }

  function collectPreference() {
    return {
      tags: splitList(elements.tagsInput.value),
      sources: splitList(elements.sourcesInput.value),
      keywords: splitList(elements.keywordsInput.value),
      weights: {
        tag: parseNumber(elements.weightTagInput.value),
        source: parseNumber(elements.weightSourceInput.value),
        keyword: parseNumber(elements.weightKeywordInput.value),
      },
    };
  }

  function splitList(value) {
    return value.split(/[\n,，]/).map((item) => item.trim()).filter(Boolean);
  }

  function parseNumber(value) {
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0;
  }

  function fillPreference(payload) {
    elements.tagsInput.value = (payload.tags || []).join(", ");
    elements.sourcesInput.value = (payload.sources || []).join(", ");
    elements.keywordsInput.value = (payload.keywords || []).join(", ");
    elements.weightTagInput.value = payload.weights ? payload.weights.tag : 0;
    elements.weightSourceInput.value = payload.weights ? payload.weights.source : 0;
    elements.weightKeywordInput.value = payload.weights ? payload.weights.keyword : 0;
  }

  function renderReport(report) {
    elements.generatedAtLabel.textContent = formatDateTime(report.generated_at);
    elements.itemCountValue.textContent = String((report.items || []).length);
    elements.displayCountValue.textContent = String(report.display_count || 0);
    elements.priorityCountValue.textContent = String(report.high_priority_count || 0);
    renderTopTitles(elements.topTitlesList, report.top_priority_items || [], "暂无重点标题");
    renderDecisions(elements.decisionList, report.decision_summary || [], "暂无决策摘要");
    elements.reportMarkdown.textContent = report.markdown || "暂无日报内容";
  }

  function renderEmptyReport() {
    elements.generatedAtLabel.textContent = "尚未生成";
    elements.itemCountValue.textContent = "-";
    elements.displayCountValue.textContent = "-";
    elements.priorityCountValue.textContent = "-";
    renderTopTitles(elements.topTitlesList, [], "暂无日报");
    renderDecisions(elements.decisionList, [], "暂无决策摘要");
    elements.reportMarkdown.textContent = "暂无日报，请先生成或选择历史记录。";
  }

  function renderSearch(result) {
    elements.searchGeneratedAtLabel.textContent = formatDateTime(result.generated_at);
    elements.searchQueryLabel.textContent = `关键词：${result.query || "-"}`;
    elements.searchItemCountValue.textContent = String((result.items || []).length);
    elements.searchDisplayCountValue.textContent = String(result.display_count || 0);
    elements.searchPriorityCountValue.textContent = String(result.high_priority_count || 0);
    renderDecisions(elements.searchDecisionList, result.decision_summary || [], "暂无搜索结果");
    elements.searchMarkdown.textContent = result.markdown || "暂无搜索结果";
  }

  function renderEmptySearch() {
    elements.searchGeneratedAtLabel.textContent = "尚未搜索";
    elements.searchQueryLabel.textContent = "关键词：-";
    elements.searchItemCountValue.textContent = "-";
    elements.searchDisplayCountValue.textContent = "-";
    elements.searchPriorityCountValue.textContent = "-";
    renderDecisions(elements.searchDecisionList, [], "暂无搜索结果");
    elements.searchMarkdown.textContent = "暂无搜索结果，请输入关键词后开始搜索。";
  }

  function renderTopTitles(container, titles, emptyText) {
    if (!titles.length) {
      container.innerHTML = `<li>${escapeHTML(emptyText)}</li>`;
      return;
    }
    container.innerHTML = titles.map((title) => `<li>${escapeHTML(title)}</li>`).join("");
  }

  function renderDecisions(container, items, emptyText) {
    if (!items.length) {
      container.textContent = emptyText;
      container.className = "decision-list empty-state";
      return;
    }

    container.className = "decision-list";
    container.innerHTML = items.map((item) => {
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
    }).join("");
  }

  function renderReportHistory(reports) {
    renderHistoryList(elements.historyList, reports, state.selectedReportName, "暂无历史日报", (name) => openReport(name), (report) => `
      <h3>${escapeHTML(formatName(report.name))}</h3>
      <div class="history-meta">
        <span>采集 ${escapeHTML(String(report.item_count || 0))}</span>
        <span>展示 ${escapeHTML(String(report.display_count || 0))}</span>
        <span>高优先级 ${escapeHTML(String(report.high_priority_count || 0))}</span>
      </div>
    `);
  }

  function renderSearchHistory(searches) {
    renderHistoryList(elements.searchHistoryList, searches, state.selectedSearchName, "暂无搜索记录", (name) => openSearch(name), (search) => `
      <h3>${escapeHTML(search.query || "未命名搜索")}</h3>
      <div class="history-meta">
        <span>${escapeHTML(formatName(search.name))}</span>
        <span>结果 ${escapeHTML(String(search.item_count || 0))}</span>
        <span>高优先级 ${escapeHTML(String(search.high_priority_count || 0))}</span>
      </div>
    `);
  }

  function renderHistoryList(container, items, selectedName, emptyText, onOpen, renderItem) {
    if (!items.length) {
      container.textContent = emptyText;
      container.className = "history-list empty-state";
      return;
    }

    container.className = "history-list";
    container.innerHTML = items.map((item) => `
      <button class="history-item${selectedName === item.name ? " is-active" : ""}" data-name="${escapeHTML(item.name)}" type="button">
        ${renderItem(item)}
      </button>
    `).join("");

    Array.from(container.querySelectorAll("[data-name]")).forEach((button) => {
      button.addEventListener("click", () => onOpen(button.getAttribute("data-name")));
    });
  }

  function highlightHistory(container, name) {
    Array.from(container.querySelectorAll("[data-name]")).forEach((button) => {
      button.classList.toggle("is-active", button.getAttribute("data-name") === name);
    });
  }

  async function copyText(value, button) {
    const text = (value || "").trim();
    if (!text) {
      updateCopyButton(button, "无内容", true);
      return;
    }

    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
      } else {
        const helper = document.createElement("textarea");
        helper.value = text;
        helper.setAttribute("readonly", "readonly");
        helper.style.position = "fixed";
        helper.style.opacity = "0";
        document.body.appendChild(helper);
        helper.select();
        document.execCommand("copy");
        document.body.removeChild(helper);
      }
      updateCopyButton(button, "已复制", false);
    } catch (error) {
      updateCopyButton(button, "失败", true);
    }
  }

  function updateCopyButton(button, label, isError) {
    button.textContent = label;
    button.classList.toggle("is-success", !isError);
    button.classList.toggle("is-error", Boolean(isError));
    window.clearTimeout(button._timer);
    button._timer = window.setTimeout(() => {
      button.textContent = "复制 Markdown";
      button.classList.remove("is-success", "is-error");
    }, 1600);
  }

  function formatName(name) {
    if (!name) {
      return "未命名记录";
    }
    const year = name.slice(0, 4);
    const month = name.slice(4, 6);
    const day = name.slice(6, 8);
    const hour = name.slice(9, 11);
    const minute = name.slice(11, 13);
    return `${year}-${month}-${day} ${hour}:${minute}`;
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

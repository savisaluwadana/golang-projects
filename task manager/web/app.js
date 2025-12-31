// API Base URL
const API_BASE = 'http://localhost:8080/api';

// Global State
let state = {
    projects: [],
    tasks: [],
    timeEntries: [],
    currentProject: null,
    activeTimer: null,
    currentView: 'dashboard'
};

// Utility Functions
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => toast.classList.add('show'), 10);
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString();
}

function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

function getPriorityBadge(priority) {
    const priorities = ['low', 'medium', 'high', 'urgent'];
    return `<span class="badge badge-${priorities[priority]}">${priorities[priority]}</span>`;
}

function getStatusBadge(status) {
    return `<span class="badge badge-${status}">${status.replace('_', ' ')}</span>`;
}

// API Calls
async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.message || 'API request failed');
        }
        
        return data;
    } catch (error) {
        showToast(error.message, 'error');
        throw error;
    }
}

// Data Loading Functions
async function loadProjects() {
    const data = await apiCall('/projects');
    state.projects = data.data || [];
    return state.projects;
}

async function loadTasks() {
    const data = await apiCall('/tasks');
    state.tasks = data.data || [];
    return state.tasks;
}

async function loadTimeEntries() {
    const data = await apiCall('/time');
    state.timeEntries = data.data || [];
    return state.timeEntries;
}

async function loadStats() {
    const data = await apiCall('/stats');
    return data.data || {};
}

async function loadReports() {
    const data = await apiCall('/reports');
    return data.data || {};
}

// View Management
function switchView(viewName) {
    // Update active nav
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-view="${viewName}"]`)?.classList.add('active');
    
    // Update active view
    document.querySelectorAll('.view').forEach(view => {
        view.classList.remove('active');
    });
    document.getElementById(`${viewName}-view`)?.classList.add('active');
    
    // Update topbar title
    const titles = {
        dashboard: 'Dashboard',
        projects: 'Projects',
        kanban: 'Kanban Board',
        tasks: 'All Tasks',
        time: 'Time Tracking',
        reports: 'Reports'
    };
    document.querySelector('.topbar-title').textContent = titles[viewName] || viewName;
    
    state.currentView = viewName;
    
    // Load view data
    loadViewData(viewName);
}

async function loadViewData(viewName) {
    switch(viewName) {
        case 'dashboard':
            await renderDashboard();
            break;
        case 'projects':
            await renderProjects();
            break;
        case 'kanban':
            await renderKanban();
            break;
        case 'tasks':
            await renderTasks();
            break;
        case 'time':
            await renderTimeTracking();
            break;
        case 'reports':
            await renderReports();
            break;
    }
}

// Dashboard
async function renderDashboard() {
    const stats = await loadStats();
    const projects = await loadProjects();
    const tasks = await loadTasks();
    
    // Render stats
    document.getElementById('stat-projects').textContent = stats.total_projects || 0;
    document.getElementById('stat-tasks').textContent = stats.total_tasks || 0;
    document.getElementById('stat-completed').textContent = stats.completed_tasks || 0;
    document.getElementById('stat-hours').textContent = (stats.total_hours_tracked || 0).toFixed(1);
    
    // Render recent projects
    const recentProjects = projects.slice(0, 5);
    const projectsHtml = recentProjects.map(project => {
        const taskCount = tasks.filter(t => t.project_id === project.id).length;
        return `
        <div class="project-card" style="border-left-color: ${project.color}">
            <div class="project-name">${project.name}</div>
            <div class="project-stats">
                <span>📋 ${taskCount} tasks</span>
            </div>
        </div>
    `}).join('');
    document.getElementById('recent-projects').innerHTML = projectsHtml || '<div class="empty-state">No projects yet</div>';
    
    // Render active tasks
    const activeTasks = tasks.filter(t => t.status !== 'done').slice(0, 5);
    const tasksHtml = activeTasks.map(task => `
        <div class="task-item">
            <div class="task-item-header">
                <div class="task-item-content">
                    <div class="task-item-description">${task.description}</div>
                    <div class="task-item-meta">
                        ${getPriorityBadge(task.priority)}
                        ${getStatusBadge(task.status)}
                    </div>
                </div>
            </div>
        </div>
    `).join('');
    document.getElementById('active-tasks').innerHTML = tasksHtml || '<div class="empty-state">No active tasks</div>';
}

// Projects
async function renderProjects() {
    const projects = await loadProjects();
    const tasks = await loadTasks();
    
    console.log('Projects:', projects);
    console.log('Tasks:', tasks);
    
    const html = projects.map(project => {
        const projectTasks = tasks.filter(t => t.project_id === project.id);
        const taskCount = projectTasks.length;
        console.log(`Project ${project.id} (${project.name}): ${taskCount} tasks`, projectTasks);
        return `
        <div class="project-card" style="border-left-color: ${project.color}" onclick="viewProject(${project.id})">
            <div class="project-header">
                <div>
                    <div class="project-name">${project.name}</div>
                    <div class="project-description">${project.description || ''}</div>
                </div>
            </div>
            <div class="project-stats">
                <span>📋 ${taskCount} tasks</span>
            </div>
            <div class="project-actions">
                <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editProject(${project.id})">Edit</button>
                <button class="btn btn-sm btn-primary" onclick="event.stopPropagation(); selectProject(${project.id})">View Kanban</button>
                <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteProject(${project.id})">Delete</button>
            </div>
        </div>
    `}).join('');
    
    document.getElementById('projects-list').innerHTML = html || '<div class="empty-state"><div class="empty-state-icon">📁</div><p>No projects yet. Create your first project!</p></div>';
}

function viewProject(projectId) {
    selectProject(projectId);
}

function selectProject(projectId) {
    state.currentProject = state.projects.find(p => p.id === projectId);
    switchView('kanban');
}

async function createProject() {
    const projectId = document.getElementById('project-id')?.value;
    const name = document.getElementById('project-name').value.trim();
    const description = document.getElementById('project-description').value.trim();
    const color = document.getElementById('project-color').value;
    
    if (!name) {
        showToast('Please enter a project name', 'error');
        return;
    }
    
    if (projectId) {
        // Update existing project
        await apiCall(`/projects/${projectId}`, {
            method: 'PUT',
            body: JSON.stringify({ name, description, color })
        });
        showToast('Project updated successfully');
    } else {
        // Create new project
        await apiCall('/projects', {
            method: 'POST',
            body: JSON.stringify({ name, description, color })
        });
        showToast('Project created successfully');
    }
    
    closeModal('project-modal');
    document.getElementById('project-form').reset();
    
    await loadProjects();
    if (state.currentView === 'projects') {
        await renderProjects();
    }
}

function editProject(projectId) {
    const project = state.projects.find(p => p.id === projectId);
    if (!project) return;
    
    // Set modal to edit mode
    document.getElementById('project-modal-title').textContent = 'Edit Project';
    document.getElementById('project-id').value = project.id;
    document.getElementById('project-name').value = project.name;
    document.getElementById('project-description').value = project.description || '';
    document.getElementById('project-color').value = project.color || '#6366f1';
    
    openModal('project-modal');
}

async function deleteProject(projectId) {
    if (!confirm('Are you sure you want to delete this project? All tasks will remain but will be unassigned.')) {
        return;
    }
    
    await apiCall(`/projects/${projectId}`, { method: 'DELETE' });
    showToast('Project deleted successfully');
    
    await loadProjects();
    if (state.currentView === 'projects') {
        await renderProjects();
    }
}

// Kanban
async function renderKanban() {
    const projectSelect = document.getElementById('kanban-project-filter');
    
    // Populate project filter
    const projects = await loadProjects();
    projectSelect.innerHTML = '<option value="">All Projects</option>' + 
        projects.map(p => `<option value="${p.id}" ${state.currentProject?.id === p.id ? 'selected' : ''}>${p.name}</option>`).join('');
    
    const projectId = state.currentProject?.id || projectSelect.value;
    
    if (!projectId) {
        document.querySelectorAll('.column-content').forEach(col => {
            col.innerHTML = '<div class="empty-state">Select a project</div>';
        });
        return;
    }
    
    const data = await apiCall(`/kanban?project_id=${projectId}`);
    const kanbanData = data.data || { backlog: [], todo: [], in_progress: [], in_review: [], done: [] };
    
    // Render columns
    ['backlog', 'todo', 'in_progress', 'in_review', 'done'].forEach(status => {
        const tasks = kanbanData[status] || [];
        const columnContent = document.getElementById(`${status}-column`);
        const countEl = columnContent.previousElementSibling.querySelector('.count');
        
        countEl.textContent = tasks.length;
        
        const html = tasks.map(task => `
            <div class="kanban-task" draggable="true" data-task-id="${task.id}" data-status="${status}" onclick="viewTask(${task.id})" style="cursor: pointer;">
                <div class="kanban-task-header">
                    <div style="flex: 1;">
                        <strong>${task.description}</strong>
                    </div>
                    ${getPriorityBadge(task.priority)}
                </div>
                <div class="kanban-task-meta">
                    ${task.category ? `<span>🏷️ ${task.category}</span>` : ''}
                    ${task.assignee ? `<span>👤 ${task.assignee}</span>` : ''}
                    ${task.due_date ? `<span>📅 ${formatDate(task.due_date)}</span>` : ''}
                    ${task.estimated_hours ? `<span>⏱️ ${task.estimated_hours}h</span>` : ''}
                </div>
                <div class="kanban-task-footer">
                    <div style="font-size: 0.75rem; color: var(--text-muted);">ID: ${task.id}</div>
                </div>
            </div>
        `).join('');
        
        columnContent.innerHTML = html || '<div class="empty-state" style="padding: 20px;">No tasks</div>';
    });
    
    // Setup drag and drop
    setupDragAndDrop();
}

function setupDragAndDrop() {
    const tasks = document.querySelectorAll('.kanban-task');
    const columns = document.querySelectorAll('.column-content');
    
    tasks.forEach(task => {
        task.addEventListener('dragstart', handleDragStart);
        task.addEventListener('dragend', handleDragEnd);
    });
    
    columns.forEach(column => {
        column.addEventListener('dragover', handleDragOver);
        column.addEventListener('drop', handleDrop);
        column.addEventListener('dragleave', handleDragLeave);
    });
}

function handleDragStart(e) {
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', this.innerHTML);
    e.dataTransfer.setData('task-id', this.dataset.taskId);
    e.dataTransfer.setData('old-status', this.dataset.status);
    this.classList.add('dragging');
}

function handleDragEnd(e) {
    this.classList.remove('dragging');
}

function handleDragOver(e) {
    if (e.preventDefault) {
        e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';
    this.style.background = 'rgba(0, 0, 0, 0.05)';
    return false;
}

function handleDragLeave(e) {
    this.style.background = '';
}

async function handleDrop(e) {
    if (e.stopPropagation) {
        e.stopPropagation();
    }
    e.preventDefault();
    
    this.style.background = '';
    
    const taskId = parseInt(e.dataTransfer.getData('task-id'));
    const oldStatus = e.dataTransfer.getData('old-status');
    const newStatus = this.id.replace('-column', '');
    
    if (oldStatus !== newStatus) {
        try {
            await apiCall('/kanban/move', {
                method: 'PUT',
                body: JSON.stringify({
                    task_id: taskId,
                    new_status: newStatus,
                    position: 0
                })
            });
            
            showToast('Task moved successfully');
            await renderKanban();
        } catch (error) {
            showToast('Failed to move task', 'error');
        }
    }
    
    return false;
}

// Tasks
async function renderTasks() {
    const tasks = await loadTasks();
    const projects = await loadProjects();
    
    // Populate filters
    const projectFilter = document.getElementById('filter-project');
    projectFilter.innerHTML = '<option value="">All Projects</option>' +
        projects.map(p => `<option value="${p.id}">${p.name}</option>`).join('');
    
    renderTasksList(tasks);
}

function renderTasksList(tasks) {
    const html = tasks.map(task => {
        const project = state.projects.find(p => p.id === task.project_id);
        return `
            <div class="task-item" style="cursor: pointer;" onclick="viewTask(${task.id})">
                <div class="task-item-header">
                    <div class="task-item-content">
                        <div class="task-item-description">${task.description}</div>
                        <div class="task-item-meta">
                            ${getPriorityBadge(task.priority)}
                            ${getStatusBadge(task.status)}
                            ${task.category ? `<span>🏷️ ${task.category}</span>` : ''}
                            ${project ? `<span>📁 ${project.name}</span>` : ''}
                            ${task.assignee ? `<span>👤 ${task.assignee}</span>` : ''}
                            ${task.due_date ? `<span>📅 ${formatDate(task.due_date)}</span>` : ''}
                            ${task.estimated_hours ? `<span>⏱️ ${task.estimated_hours}h</span>` : ''}
                            ${task.tags && task.tags.length ? `<span>🏷️ ${task.tags.join(', ')}</span>` : ''}
                        </div>
                    </div>
                    <div class="task-item-actions">
                        <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editTask(${task.id}); return false;">Edit</button>
                        <button class="btn btn-sm btn-success" onclick="event.stopPropagation(); markTaskDone(${task.id}); return false;" ${task.status === 'done' ? 'disabled' : ''}>
                            ${task.status === 'done' ? '✓ Done' : 'Mark Done'}
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteTask(${task.id}); return false;">Delete</button>
                    </div>
                </div>
            </div>
        `;
    }).join('');
    
    document.getElementById('tasks-list').innerHTML = html || '<div class="empty-state"><div class="empty-state-icon">📝</div><p>No tasks found</p></div>';
}

function filterTasks() {
    const projectId = document.getElementById('filter-project').value;
    const status = document.getElementById('filter-status').value;
    const priority = document.getElementById('filter-priority').value;
    const search = document.getElementById('search-tasks').value.toLowerCase();
    
    let filtered = state.tasks;
    
    if (projectId) {
        filtered = filtered.filter(t => t.project_id === parseInt(projectId));
    }
    
    if (status) {
        filtered = filtered.filter(t => t.status === status);
    }
    
    if (priority !== '') {
        filtered = filtered.filter(t => t.priority === parseInt(priority));
    }
    
    if (search) {
        filtered = filtered.filter(t => 
            t.description.toLowerCase().includes(search) ||
            (t.category && t.category.toLowerCase().includes(search)) ||
            (t.assignee && t.assignee.toLowerCase().includes(search))
        );
    }
    
    renderTasksList(filtered);
}

function openTaskModal(taskId = null) {
    console.log('openTaskModal called with taskId:', taskId);
    const modal = document.getElementById('task-modal');
    const title = document.getElementById('task-modal-title');
    const form = document.getElementById('task-form');
    const commentsSection = document.getElementById('comments-section');
    
    if (!modal || !title || !form) {
        console.error('Required modal elements not found:', { modal, title, form });
        showToast('Error: Modal elements not found', 'error');
        return;
    }
    
    if (taskId) {
        // Edit mode
        const task = state.tasks.find(t => t.id === taskId);
        if (!task) {
            console.error('Task not found:', taskId);
            return;
        }
        
        title.textContent = 'Edit Task';
        document.getElementById('task-id').value = task.id;
        document.getElementById('task-description').value = task.description;
        document.getElementById('task-project').value = task.project_id || '';
        document.getElementById('task-status').value = task.status;
        document.getElementById('task-priority').value = task.priority;
        document.getElementById('task-category').value = task.category || '';
        document.getElementById('task-assignee').value = task.assignee || '';
        document.getElementById('task-due-date').value = task.due_date ? task.due_date.split('T')[0] : '';
        document.getElementById('task-estimated-hours').value = task.estimated_hours || '';
        document.getElementById('task-tags').value = task.tags ? task.tags.join(', ') : '';
        
        // Show and load comments
        if (commentsSection) {
            commentsSection.style.display = 'block';
            loadComments(taskId);
        }
    } else {
        // Create mode
        title.textContent = 'Create Task';
        form.reset();
        document.getElementById('task-id').value = '';
        if (commentsSection) {
            commentsSection.style.display = 'none';
        }
        if (state.currentProject) {
            document.getElementById('task-project').value = state.currentProject.id;
        }
    }
    
    // Populate project dropdown
    const projectSelect = document.getElementById('task-project');
    if (projectSelect) {
        projectSelect.innerHTML = '<option value="">No Project</option>' +
            state.projects.map(p => `<option value="${p.id}">${p.name}</option>`).join('');
        
        if (taskId) {
            const task = state.tasks.find(t => t.id === taskId);
            if (task && task.project_id) {
                projectSelect.value = task.project_id;
            }
        } else if (state.currentProject) {
            projectSelect.value = state.currentProject.id;
        }
    }
    
    openModal('task-modal');
}

function viewTask(taskId) {
    const task = state.tasks.find(t => t.id === taskId);
    if (!task) return;
    
    const project = state.projects.find(p => p.id === task.project_id);
    
    // Populate view modal
    document.getElementById('view-task-description').textContent = task.description;
    document.getElementById('view-task-project').innerHTML = project ? `<span class="badge">${project.name}</span>` : '-';
    document.getElementById('view-task-status').innerHTML = getStatusBadge(task.status);
    document.getElementById('view-task-priority').innerHTML = getPriorityBadge(task.priority);
    document.getElementById('view-task-assignee').textContent = task.assignee || '-';
    document.getElementById('view-task-due-date').textContent = task.due_date ? formatDate(task.due_date) : '-';
    document.getElementById('view-task-hours').textContent = task.estimated_hours ? `${task.estimated_hours}h` : '-';
    document.getElementById('view-task-category').textContent = task.category || '-';
    document.getElementById('view-task-tags').textContent = task.tags && task.tags.length ? task.tags.join(', ') : '-';
    
    // Store task ID for editing and comments
    document.getElementById('view-task-modal').dataset.taskId = taskId;
    
    // Load comments
    loadViewComments(taskId);
    
    openModal('view-task-modal');
}

function editTask(taskId) {
    console.log('editTask called with taskId:', taskId);
    if (!taskId) {
        console.error('No task ID provided to editTask');
        showToast('Error: No task ID provided', 'error');
        return;
    }
    openTaskModal(taskId);
}

async function saveTask() {
    const taskId = document.getElementById('task-id').value;
    const description = document.getElementById('task-description').value.trim();
    const projectId = document.getElementById('task-project').value;
    const status = document.getElementById('task-status').value;
    const priority = document.getElementById('task-priority').value;
    const category = document.getElementById('task-category').value.trim();
    const assignee = document.getElementById('task-assignee').value.trim();
    const dueDate = document.getElementById('task-due-date').value;
    const estimatedHours = document.getElementById('task-estimated-hours').value;
    const tagsStr = document.getElementById('task-tags').value;
    
    if (!description) {
        showToast('Please enter a task description', 'error');
        return;
    }
    
    const taskData = {
        description,
        priority,
        status,
        category: category || undefined,
        assignee: assignee || undefined,
        due_date: dueDate || undefined,
        estimated_hours: estimatedHours ? parseFloat(estimatedHours) : undefined,
        project_id: projectId ? parseInt(projectId) : undefined,
        tags: tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(t => t) : undefined
    };
    
    if (taskId) {
        // Update
        await apiCall(`/tasks/${taskId}`, {
            method: 'PUT',
            body: JSON.stringify(taskData)
        });
        showToast('Task updated successfully');
    } else {
        // Create
        await apiCall('/tasks', {
            method: 'POST',
            body: JSON.stringify(taskData)
        });
        showToast('Task created successfully');
    }
    
    closeModal('task-modal');
    document.getElementById('task-form').reset();
    
    await loadTasks();
    loadViewData(state.currentView);
}

async function markTaskDone(taskId) {
    await apiCall(`/tasks/${taskId}/done`, { method: 'PUT' });
    showToast('Task marked as done');
    
    await loadTasks();
    loadViewData(state.currentView);
}

async function deleteTask(taskId) {
    if (!confirm('Are you sure you want to delete this task?')) {
        return;
    }
    
    await apiCall(`/tasks/${taskId}`, { method: 'DELETE' });
    showToast('Task deleted successfully');
    
    await loadTasks();
    loadViewData(state.currentView);
}

// Time Tracking
async function renderTimeTracking() {
    const entries = await loadTimeEntries();
    const tasks = await loadTasks();
    
    // Populate task dropdown
    const taskSelect = document.getElementById('time-task-select');
    taskSelect.innerHTML = '<option value="">Select a task...</option>' +
        tasks.filter(t => !t.completed).map(t => `<option value="${t.id}">${t.description}</option>`).join('');
    
    // Render time entries
    const html = entries.map(entry => {
        const task = tasks.find(t => t.id === entry.task_id);
        const duration = entry.end_time ? 
            Math.floor((new Date(entry.end_time) - new Date(entry.start_time)) / 1000) :
            Math.floor((new Date() - new Date(entry.start_time)) / 1000);
        
        return `
            <div class="time-entry">
                <div class="time-entry-header">
                    <div>
                        <div class="time-entry-task">${task ? task.description : 'Unknown Task'}</div>
                        ${entry.note ? `<div class="time-entry-notes"><strong>Working on:</strong> ${entry.note}</div>` : ''}
                        <div class="time-entry-meta">
                            ${formatDate(entry.start_time)} - ${entry.end_time ? formatDate(entry.end_time) : 'In Progress'}
                        </div>
                    </div>
                    <div class="time-entry-duration">${formatDuration(duration)}</div>
                </div>
            </div>
        `;
    }).join('');
    
    document.getElementById('time-entries-list').innerHTML = html || '<div class="empty-state">No time entries yet</div>';
    
    // Check for active timer
    checkActiveTimer();
}

async function checkActiveTimer() {
    const entries = await loadTimeEntries();
    const activeEntry = entries.find(e => !e.end_time);
    
    if (activeEntry) {
        state.activeTimer = activeEntry;
        showActiveTimer();
        startTimerUpdate();
    } else {
        state.activeTimer = null;
        hideActiveTimer();
    }
}

function showActiveTimer() {
    const widget = document.getElementById('active-timer-widget');
    const taskName = document.getElementById('active-timer-task');
    const task = state.tasks.find(t => t.id === state.activeTimer.task_id);
    
    taskName.textContent = task ? task.description : 'Unknown Task';
    widget.style.display = 'block';
    
    updateTimerDisplay();
}

function hideActiveTimer() {
    document.getElementById('active-timer-widget').style.display = 'none';
}

function updateTimerDisplay() {
    if (!state.activeTimer) return;
    
    const elapsed = Math.floor((new Date() - new Date(state.activeTimer.start_time)) / 1000);
    document.getElementById('active-timer-duration').textContent = formatDuration(elapsed);
}

function startTimerUpdate() {
    setInterval(updateTimerDisplay, 1000);
}

async function startTimer() {
    const taskId = document.getElementById('time-task-select').value;
    const notes = document.getElementById('time-notes').value.trim();
    
    if (!taskId) {
        showToast('Please select a task', 'error');
        return;
    }
    
    if (state.activeTimer) {
        showToast('Please stop the current timer first', 'error');
        return;
    }
    
    await apiCall('/time/start', {
        method: 'POST',
        body: JSON.stringify({
            task_id: parseInt(taskId),
            note: notes || undefined
        })
    });
    
    showToast('Timer started');
    document.getElementById('time-notes').value = '';
    
    await renderTimeTracking();
}

async function stopTimer() {
    if (!state.activeTimer) return;
    
    await apiCall(`/time/${state.activeTimer.id}/stop`, {
        method: 'PUT'
    });
    
    showToast('Timer stopped');
    state.activeTimer = null;
    
    await renderTimeTracking();
}

// Comments
async function loadComments(taskId) {
    const comments = await apiCall(`/comments?task_id=${taskId}`);
    renderComments(comments || []);
}

function renderComments(comments) {
    const commentsList = document.getElementById('comments-list');
    
    if (!comments || comments.length === 0) {
        commentsList.innerHTML = '<div style="color: var(--text-muted); font-size: 0.875rem;">No comments yet</div>';
        return;
    }
    
    commentsList.innerHTML = comments.map(comment => {
        const date = new Date(comment.created_at);
        const dateStr = date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
        
        return `
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${comment.author}</span>
                    <span class="comment-date">${dateStr}</span>
                </div>
                <div class="comment-text">${comment.text}</div>
            </div>
        `;
    }).join('');
}

async function addComment() {
    const taskId = document.getElementById('task-id').value;
    const author = document.getElementById('comment-author').value.trim();
    const text = document.getElementById('comment-text').value.trim();
    
    if (!taskId || !author || !text) {
        showToast('Please fill in all comment fields');
        return;
    }
    
    await apiCall('/comments', {
        method: 'POST',
        body: JSON.stringify({
            task_id: parseInt(taskId),
            author: author,
            text: text
        })
    });
    
    // Clear inputs
    document.getElementById('comment-text').value = '';
    
    // Reload comments
    await loadComments(parseInt(taskId));
    showToast('Comment added');
}

// Reports
async function renderReports() {
    const reports = await loadReports();
    
    // Project completion
    if (reports.project_completion && reports.project_completion.length) {
        const html = reports.project_completion.map(pc => {
            const percentage = pc.total_tasks > 0 ? Math.round((pc.completed_tasks / pc.total_tasks) * 100) : 0;
            return `
                <div class="progress-label">
                    <span>${pc.project_name}</span>
                    <span>${pc.completed_tasks}/${pc.total_tasks} tasks</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${percentage}%">${percentage}%</div>
                </div>
            `;
        }).join('');
        document.getElementById('project-completion-report').innerHTML = html;
    }
    
    // Time by project
    if (reports.time_by_project && reports.time_by_project.length) {
        const html = reports.time_by_project.map(tp => `
            <div class="report-item">
                <div class="report-value">${tp.total_hours.toFixed(1)}h</div>
                <div class="report-label">${tp.project_name}</div>
            </div>
        `).join('');
        document.getElementById('time-by-project-report').innerHTML = html;
    }
    
    // Time by user
    if (reports.time_by_user && reports.time_by_user.length) {
        const html = reports.time_by_user.map(tu => `
            <div class="report-item">
                <div class="report-value">${tu.total_hours.toFixed(1)}h</div>
                <div class="report-label">${tu.assignee || 'Unassigned'}</div>
            </div>
        `).join('');
        document.getElementById('time-by-user-report').innerHTML = html;
    }
    
    // Task distribution
    if (reports.task_distribution) {
        const dist = reports.task_distribution;
        const total = dist.backlog + dist.todo + dist.in_progress + dist.in_review + dist.done;
        
        const html = Object.entries(dist).map(([status, count]) => {
            const percentage = total > 0 ? Math.round((count / total) * 100) : 0;
            return `
                <div class="progress-label">
                    <span>${status.replace('_', ' ').toUpperCase()}</span>
                    <span>${count} tasks</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${percentage}%">${percentage}%</div>
                </div>
            `;
        }).join('');
        document.getElementById('task-distribution-report').innerHTML = html;
    }
}

// Comments
async function loadComments(taskId) {
    try {
        const response = await apiCall(`/comments?task_id=${taskId}`);
        const comments = response.data || [];
        renderComments(comments);
    } catch (error) {
        console.error('Error loading comments:', error);
        renderComments([]);
    }
}

async function loadViewComments(taskId) {
    try {
        const response = await apiCall(`/comments?task_id=${taskId}`);
        const comments = response.data || [];
        renderViewComments(comments);
    } catch (error) {
        console.error('Error loading comments:', error);
        renderViewComments([]);
    }
}

function renderComments(comments) {
    const commentsList = document.getElementById('comments-list');
    
    if (!comments || comments.length === 0) {
        commentsList.innerHTML = '<p style="color: var(--text-muted); font-size: 0.875rem;">No comments yet</p>';
        return;
    }
    
    commentsList.innerHTML = comments.map(comment => {
        const date = new Date(comment.created_at);
        const formattedDate = date.toLocaleString();
        
        return `
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${escapeHtml(comment.author)}</span>
                    <span class="comment-date">${formattedDate}</span>
                </div>
                <div class="comment-text">${escapeHtml(comment.text)}</div>
            </div>
        `;
    }).join('');
}

function renderViewComments(comments) {
    const commentsList = document.getElementById('view-comments-list');
    
    if (!comments || comments.length === 0) {
        commentsList.innerHTML = '<p style="color: var(--text-muted); font-size: 0.875rem;">No comments yet</p>';
        return;
    }
    
    commentsList.innerHTML = comments.map(comment => {
        const date = new Date(comment.created_at);
        const formattedDate = date.toLocaleString();
        
        return `
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${escapeHtml(comment.author)}</span>
                    <span class="comment-date">${formattedDate}</span>
                </div>
                <div class="comment-text">${escapeHtml(comment.text)}</div>
            </div>
        `;
    }).join('');
}

async function addComment() {
    const taskId = document.getElementById('task-id').value;
    const author = document.getElementById('comment-author').value.trim();
    const text = document.getElementById('comment-text').value.trim();
    
    if (!taskId) {
        showToast('Please save the task first before adding comments', 'error');
        return;
    }
    
    if (!text) {
        showToast('Please enter a comment', 'error');
        return;
    }
    
    try {
        await apiCall('/comments', {
            method: 'POST',
            body: JSON.stringify({
                task_id: parseInt(taskId),
                author: author || 'Anonymous',
                text: text
            })
        });
        
        // Clear inputs
        document.getElementById('comment-text').value = '';
        
        // Reload comments
        await loadComments(taskId);
        showToast('Comment added');
    } catch (error) {
        showToast('Failed to add comment', 'error');
    }
}

async function addViewComment() {
    const modal = document.getElementById('view-task-modal');
    const taskId = modal.dataset.taskId;
    const author = document.getElementById('view-comment-author').value.trim();
    const text = document.getElementById('view-comment-text').value.trim();
    
    if (!text) {
        showToast('Please enter a comment', 'error');
        return;
    }
    
    try {
        await apiCall('/comments', {
            method: 'POST',
            body: JSON.stringify({
                task_id: parseInt(taskId),
                author: author || 'Anonymous',
                text: text
            })
        });
        
        // Clear inputs
        document.getElementById('view-comment-text').value = '';
        
        // Reload comments
        await loadViewComments(taskId);
        showToast('Comment added');
    } catch (error) {
        showToast('Failed to add comment', 'error');
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Modal Functions
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('active');
    } else {
        console.error(`Modal with id "${modalId}" not found`);
    }
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('active');
    }
}

// Sidebar Toggle (Mobile)
function toggleSidebar() {
    document.querySelector('.sidebar').classList.toggle('active');
}

// Event Listeners
document.addEventListener('DOMContentLoaded', async () => {
    // Navigation
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const view = item.dataset.view;
            switchView(view);
        });
    });
    
    // Sidebar toggle
    document.getElementById('btn-toggle-sidebar')?.addEventListener('click', toggleSidebar);
    
    // Project modal
    document.getElementById('btn-create-project')?.addEventListener('click', () => {
        // Reset form for creating new project
        document.getElementById('project-modal-title').textContent = 'Create New Project';
        document.getElementById('project-form').reset();
        document.getElementById('project-id').value = '';
        document.getElementById('project-color').value = '#000000';
        openModal('project-modal');
    });
    document.getElementById('btn-save-project')?.addEventListener('click', createProject);
    
    // Task modal - handle both create task buttons
    const btnCreateTask = document.getElementById('btn-create-task');
    const btnCreateTaskAlt = document.querySelector('.btn-create-task-alt');
    console.log('Task buttons found:', btnCreateTask, btnCreateTaskAlt);
    
    btnCreateTask?.addEventListener('click', () => {
        console.log('btn-create-task clicked');
        openTaskModal();
    });
    btnCreateTaskAlt?.addEventListener('click', () => {
        console.log('btn-create-task-alt clicked');
        openTaskModal();
    });
    document.getElementById('btn-save-task')?.addEventListener('click', saveTask);
    
    // Close modal buttons
    document.querySelectorAll('.close-btn, .btn-secondary').forEach(btn => {
        btn.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                modal.classList.remove('active');
            }
        });
    });
    
    // Kanban filters
    document.getElementById('kanban-project-filter')?.addEventListener('change', () => {
        state.currentProject = state.projects.find(p => p.id === parseInt(document.getElementById('kanban-project-filter').value));
        renderKanban();
    });
    
    // Task filters
    document.getElementById('filter-project')?.addEventListener('change', filterTasks);
    document.getElementById('filter-status')?.addEventListener('change', filterTasks);
    document.getElementById('filter-priority')?.addEventListener('change', filterTasks);
    document.getElementById('search-tasks')?.addEventListener('input', filterTasks);
    
    // Time tracking
    document.getElementById('btn-start-timer')?.addEventListener('click', startTimer);
    document.getElementById('btn-stop-timer')?.addEventListener('click', stopTimer);
    document.getElementById('time-task-select')?.addEventListener('change', function() {
        const taskId = parseInt(this.value);
        const descEl = document.getElementById('selected-task-desc');
        if (taskId && descEl) {
            const task = state.tasks.find(t => t.id === taskId);
            if (task) {
                descEl.innerHTML = `<strong>Task:</strong> ${task.description}`;
                descEl.style.display = 'block';
            } else {
                descEl.style.display = 'none';
            }
        } else if (descEl) {
            descEl.style.display = 'none';
        }
    });
    
    // Comments
    document.getElementById('btn-add-comment')?.addEventListener('click', addComment);
    document.getElementById('btn-add-view-comment')?.addEventListener('click', addViewComment);
    
    // Edit from view modal
    document.getElementById('btn-edit-from-view')?.addEventListener('click', () => {
        const taskId = document.getElementById('view-task-modal').dataset.taskId;
        closeModal('view-task-modal');
        editTask(parseInt(taskId));
    });
    
    // Initial load
    await loadProjects();
    await loadTasks();
    switchView('dashboard');
    
    // Check for active timer
    await checkActiveTimer();
});

const API_BASE = 'http://localhost:8080/api';

let state = {
    students: [],
    courses: [],
    enrollments: [],
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
        
        if (!data.success) {
            throw new Error(data.message || 'API request failed');
        }
        
        return data;
    } catch (error) {
        showToast(error.message, 'error');
        throw error;
    }
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString();
}

function getStatusBadge(status) {
    return `<span class="badge badge-${status}">${status}</span>`;
}

// Navigation
function switchView(viewName) {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-view="${viewName}"]`)?.classList.add('active');
    
    document.querySelectorAll('.view').forEach(view => {
        view.classList.remove('active');
    });
    document.getElementById(`${viewName}-view`)?.classList.add('active');
    
    const titles = {
        dashboard: 'Dashboard',
        students: 'Students',
        courses: 'Courses',
        enrollments: 'Enrollments'
    };
    document.getElementById('view-title').textContent = titles[viewName] || viewName;
    
    state.currentView = viewName;
    loadViewData(viewName);
}

async function loadViewData(viewName) {
    switch(viewName) {
        case 'dashboard':
            await renderDashboard();
            break;
        case 'students':
            await renderStudents();
            break;
        case 'courses':
            await renderCourses();
            break;
        case 'enrollments':
            await renderEnrollments();
            break;
    }
}

// Dashboard
async function renderDashboard() {
    const statsData = await apiCall('/stats');
    const stats = statsData.data;
    
    document.getElementById('stat-total-students').textContent = stats.total_students || 0;
    document.getElementById('stat-active-students').textContent = stats.active_students || 0;
    document.getElementById('stat-total-courses').textContent = stats.total_courses || 0;
    document.getElementById('stat-enrollments').textContent = stats.total_enrollments || 0;
    
    // Load recent students
    const studentsData = await apiCall('/students');
    const students = studentsData.data || [];
    const recentStudents = students.slice(0, 5);
    
    const html = `
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Enrollment Date</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                ${recentStudents.map(s => `
                    <tr>
                        <td>${s.first_name} ${s.last_name}</td>
                        <td>${s.email}</td>
                        <td>${formatDate(s.enrollment_date)}</td>
                        <td>${getStatusBadge(s.status)}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    
    document.getElementById('recent-students').innerHTML = html || '<div class="empty-state">No students yet</div>';
}

// Students
async function renderStudents() {
    const data = await apiCall('/students');
    state.students = data.data || [];
    
    const html = `
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Phone</th>
                    <th>Enrollment Date</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${state.students.map(s => `
                    <tr>
                        <td>${s.id}</td>
                        <td>${s.first_name} ${s.last_name}</td>
                        <td>${s.email}</td>
                        <td>${s.phone || '-'}</td>
                        <td>${formatDate(s.enrollment_date)}</td>
                        <td>${getStatusBadge(s.status)}</td>
                        <td class="actions">
                            <button class="btn btn-sm btn-secondary" onclick="editStudent(${s.id})">Edit</button>
                            <button class="btn btn-sm btn-danger" onclick="deleteStudent(${s.id})">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    
    document.getElementById('students-list').innerHTML = html || '<div class="empty-state">No students yet</div>';
}

function openStudentModal(studentId = null) {
    const modal = document.getElementById('student-modal');
    const title = document.getElementById('student-modal-title');
    const form = document.getElementById('student-form');
    
    if (studentId) {
        const student = state.students.find(s => s.id === studentId);
        if (!student) return;
        
        title.textContent = 'Edit Student';
        document.getElementById('student-id').value = student.id;
        document.getElementById('student-first-name').value = student.first_name;
        document.getElementById('student-last-name').value = student.last_name;
        document.getElementById('student-email').value = student.email;
        document.getElementById('student-phone').value = student.phone || '';
        document.getElementById('student-dob').value = student.date_of_birth || '';
        document.getElementById('student-enrollment-date').value = student.enrollment_date || '';
        document.getElementById('student-address').value = student.address || '';
        document.getElementById('student-status').value = student.status;
    } else {
        title.textContent = 'Add Student';
        form.reset();
        document.getElementById('student-id').value = '';
        document.getElementById('student-status').value = 'active';
    }
    
    modal.classList.add('active');
}

function editStudent(studentId) {
    openStudentModal(studentId);
}

async function saveStudent(e) {
    if (e) e.preventDefault();
    
    const studentId = document.getElementById('student-id').value;
    const studentData = {
        first_name: document.getElementById('student-first-name').value.trim(),
        last_name: document.getElementById('student-last-name').value.trim(),
        email: document.getElementById('student-email').value.trim(),
        phone: document.getElementById('student-phone').value.trim(),
        date_of_birth: document.getElementById('student-dob').value,
        enrollment_date: document.getElementById('student-enrollment-date').value,
        address: document.getElementById('student-address').value.trim(),
        status: document.getElementById('student-status').value
    };
    
    if (!studentData.first_name || !studentData.last_name || !studentData.email) {
        showToast('Please fill in all required fields', 'error');
        return;
    }
    
    try {
    
        if (studentId) {
            studentData.id = parseInt(studentId);
            await apiCall('/students', {
                method: 'PUT',
                body: JSON.stringify(studentData)
            });
            showToast('Student updated successfully');
        } else {
            await apiCall('/students', {
                method: 'POST',
                body: JSON.stringify(studentData)
            });
            showToast('Student added successfully');
        }
        
        closeModal('student-modal');
        await renderStudents();
    } catch (error) {
        console.error('Error saving student:', error);
    }
}

async function deleteStudent(studentId) {
    if (!confirm('Are you sure you want to delete this student?')) return;
    
    await apiCall(`/students?id=${studentId}`, { method: 'DELETE' });
    showToast('Student deleted successfully');
    await renderStudents();
}

// Courses
async function renderCourses() {
    const data = await apiCall('/courses');
    state.courses = data.data || [];
    
    const html = `
        <table>
            <thead>
                <tr>
                    <th>Code</th>
                    <th>Name</th>
                    <th>Credits</th>
                    <th>Instructor</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${state.courses.map(c => `
                    <tr>
                        <td><strong>${c.code}</strong></td>
                        <td>${c.name}</td>
                        <td>${c.credits}</td>
                        <td>${c.instructor || '-'}</td>
                        <td class="actions">
                            <button class="btn btn-sm btn-secondary" onclick="editCourse(${c.id})">Edit</button>
                            <button class="btn btn-sm btn-danger" onclick="deleteCourse(${c.id})">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    
    document.getElementById('courses-list').innerHTML = html || '<div class="empty-state">No courses yet</div>';
}

function openCourseModal(courseId = null) {
    const modal = document.getElementById('course-modal');
    const title = document.getElementById('course-modal-title');
    const form = document.getElementById('course-form');
    
    if (courseId) {
        const course = state.courses.find(c => c.id === courseId);
        if (!course) return;
        
        title.textContent = 'Edit Course';
        document.getElementById('course-id').value = course.id;
        document.getElementById('course-code').value = course.code;
        document.getElementById('course-name').value = course.name;
        document.getElementById('course-description').value = course.description || '';
        document.getElementById('course-credits').value = course.credits;
        document.getElementById('course-instructor').value = course.instructor || '';
    } else {
        title.textContent = 'Add Course';
        form.reset();
        document.getElementById('course-id').value = '';
        document.getElementById('course-credits').value = 3;
    }
    
    modal.classList.add('active');
}

function editCourse(courseId) {
    openCourseModal(courseId);
}

async function saveCourse(e) {
    if (e) e.preventDefault();
    
    const courseId = document.getElementById('course-id').value;
    const courseData = {
        code: document.getElementById('course-code').value.trim(),
        name: document.getElementById('course-name').value.trim(),
        description: document.getElementById('course-description').value.trim(),
        credits: parseInt(document.getElementById('course-credits').value),
        instructor: document.getElementById('course-instructor').value.trim()
    };
    
    if (!courseData.code || !courseData.name) {
        showToast('Please fill in all required fields', 'error');
        return;
    }
    
    try {
    
        if (courseId) {
            courseData.id = parseInt(courseId);
            await apiCall('/courses', {
                method: 'PUT',
                body: JSON.stringify(courseData)
            });
            showToast('Course updated successfully');
        } else {
            await apiCall('/courses', {
                method: 'POST',
                body: JSON.stringify(courseData)
            });
            showToast('Course added successfully');
        }
        
        closeModal('course-modal');
        await renderCourses();
    } catch (error) {
        console.error('Error saving course:', error);
    }
}

async function deleteCourse(courseId) {
    if (!confirm('Are you sure you want to delete this course?')) return;
    
    await apiCall(`/courses?id=${courseId}`, { method: 'DELETE' });
    showToast('Course deleted successfully');
    await renderCourses();
}

// Enrollments
async function renderEnrollments() {
    const data = await apiCall('/enrollments');
    state.enrollments = data.data || [];
    
    // Also load students and courses for display
    if (state.students.length === 0) {
        const studentsData = await apiCall('/students');
        state.students = studentsData.data || [];
    }
    if (state.courses.length === 0) {
        const coursesData = await apiCall('/courses');
        state.courses = coursesData.data || [];
    }
    
    const html = `
        <table>
            <thead>
                <tr>
                    <th>Student</th>
                    <th>Course</th>
                    <th>Enrollment Date</th>
                    <th>Grade</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${state.enrollments.map(e => {
                    const student = state.students.find(s => s.id === e.student_id);
                    const course = state.courses.find(c => c.id === e.course_id);
                    return `
                        <tr>
                            <td>${student ? `${student.first_name} ${student.last_name}` : 'Unknown'}</td>
                            <td>${course ? `${course.code} - ${course.name}` : 'Unknown'}</td>
                            <td>${formatDate(e.enrollment_date)}</td>
                            <td>${e.grade || '-'}</td>
                            <td>${getStatusBadge(e.status)}</td>
                            <td class="actions">
                                <button class="btn btn-sm btn-secondary" onclick="editEnrollment(${e.id})">Edit</button>
                            </td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
    
    document.getElementById('enrollments-list').innerHTML = html || '<div class="empty-state">No enrollments yet</div>';
}

async function openEnrollmentModal(enrollmentId = null) {
    const modal = document.getElementById('enrollment-modal');
    const title = document.getElementById('enrollment-modal-title');
    const form = document.getElementById('enrollment-form');
    
    // Load students and courses for dropdowns
    if (state.students.length === 0) {
        const studentsData = await apiCall('/students');
        state.students = studentsData.data || [];
    }
    if (state.courses.length === 0) {
        const coursesData = await apiCall('/courses');
        state.courses = coursesData.data || [];
    }
    
    // Populate student dropdown
    const studentSelect = document.getElementById('enrollment-student');
    studentSelect.innerHTML = '<option value="">Select student...</option>' +
        state.students.filter(s => s.status === 'active').map(s => 
            `<option value="${s.id}">${s.first_name} ${s.last_name}</option>`
        ).join('');
    
    // Populate course dropdown
    const courseSelect = document.getElementById('enrollment-course');
    courseSelect.innerHTML = '<option value="">Select course...</option>' +
        state.courses.map(c => 
            `<option value="${c.id}">${c.code} - ${c.name}</option>`
        ).join('');
    
    if (enrollmentId) {
        const enrollment = state.enrollments.find(e => e.id === enrollmentId);
        if (!enrollment) return;
        
        title.textContent = 'Edit Enrollment';
        document.getElementById('enrollment-id').value = enrollment.id;
        document.getElementById('enrollment-student').value = enrollment.student_id;
        document.getElementById('enrollment-course').value = enrollment.course_id;
        document.getElementById('enrollment-date').value = enrollment.enrollment_date || '';
        document.getElementById('enrollment-grade').value = enrollment.grade || '';
        document.getElementById('enrollment-status').value = enrollment.status;
        
        // Disable student and course selection when editing
        studentSelect.disabled = true;
        courseSelect.disabled = true;
    } else {
        title.textContent = 'New Enrollment';
        form.reset();
        document.getElementById('enrollment-id').value = '';
        document.getElementById('enrollment-status').value = 'enrolled';
        studentSelect.disabled = false;
        courseSelect.disabled = false;
    }
    
    modal.classList.add('active');
}

function editEnrollment(enrollmentId) {
    openEnrollmentModal(enrollmentId);
}

async function saveEnrollment(e) {
    if (e) e.preventDefault();
    
    const enrollmentId = document.getElementById('enrollment-id').value;
    const enrollmentData = {
        student_id: parseInt(document.getElementById('enrollment-student').value),
        course_id: parseInt(document.getElementById('enrollment-course').value),
        enrollment_date: document.getElementById('enrollment-date').value,
        grade: document.getElementById('enrollment-grade').value.trim(),
        status: document.getElementById('enrollment-status').value
    };
    
    if (!enrollmentData.student_id || !enrollmentData.course_id) {
        showToast('Please select student and course', 'error');
        return;
    }
    
    try {
    
        if (enrollmentId) {
            enrollmentData.id = parseInt(enrollmentId);
            await apiCall('/enrollments', {
                method: 'PUT',
                body: JSON.stringify(enrollmentData)
            });
            showToast('Enrollment updated successfully');
        } else {
            await apiCall('/enrollments', {
                method: 'POST',
                body: JSON.stringify(enrollmentData)
            });
            showToast('Enrollment created successfully');
        }
        
        closeModal('enrollment-modal');
        await renderEnrollments();
    } catch (error) {
        console.error('Error saving enrollment:', error);
    }
}

// Modal Functions
function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
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
    
    // Student modal
    document.getElementById('btn-add-student')?.addEventListener('click', (e) => {
        e.preventDefault();
        openStudentModal();
    });
    document.getElementById('btn-save-student')?.addEventListener('click', (e) => {
        e.preventDefault();
        saveStudent(e);
    });
    
    // Course modal
    document.getElementById('btn-add-course')?.addEventListener('click', (e) => {
        e.preventDefault();
        openCourseModal();
    });
    document.getElementById('btn-save-course')?.addEventListener('click', (e) => {
        e.preventDefault();
        saveCourse(e);
    });
    
    // Enrollment modal
    document.getElementById('btn-add-enrollment')?.addEventListener('click', (e) => {
        e.preventDefault();
        openEnrollmentModal();
    });
    document.getElementById('btn-save-enrollment')?.addEventListener('click', (e) => {
        e.preventDefault();
        saveEnrollment(e);
    });
    
    // Close modals
    document.querySelectorAll('.close-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                modal.classList.remove('active');
            }
        });
    });
    
    // Initial load
    await renderDashboard();
});

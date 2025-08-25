class TodoApp {
    constructor() {
        this.todos = [];
        this.initializeEventListeners();
        this.loadTodos();
    }

    initializeEventListeners() {
        const todoForm = document.getElementById('todoForm');
        todoForm.addEventListener('submit', (e) => this.handleAddTodo(e));
    }

    async loadTodos() {
        try {
            const response = await fetch('/api/todos');
            this.todos = await response.json();
            this.renderTodos();
        } catch (error) {
            console.error('Failed to load todos:', error);
            this.showError('Failed to load todos');
        }
    }

    async handleAddTodo(e) {
        e.preventDefault();

        const title = document.getElementById('title').value.trim();
        const description = document.getElementById('description').value.trim();

        if (!title) return;

        try {
            const response = await fetch('/api/todos', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ title, description }),
            });

            if (response.ok) {
                const newTodo = await response.json();
                this.todos.unshift(newTodo);
                this.renderTodos();

                // Clear form
                document.getElementById('title').value = '';
                document.getElementById('description').value = '';
            } else {
                throw new Error('Failed to create todo');
            }
        } catch (error) {
            console.error('Failed to add todo:', error);
            this.showError('Failed to add todo');
        }
    }

    async toggleTodo(id) {
        const todo = this.todos.find(t => t.id === id);
        if (!todo) return;

        try {
            const response = await fetch(`/api/todos/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    title: todo.title,
                    description: todo.description,
                    completed: !todo.completed,
                }),
            });

            if (response.ok) {
                const updatedTodo = await response.json();
                const index = this.todos.findIndex(t => t.id === id);
                this.todos[index] = updatedTodo;
                this.renderTodos();
            } else {
                throw new Error('Failed to update todo');
            }
        } catch (error) {
            console.error('Failed to toggle todo:', error);
            this.showError('Failed to update todo');
        }
    }

    async deleteTodo(id) {
        if (!confirm('Are you sure you want to delete this todo?')) return;

        try {
            const response = await fetch(`/api/todos/${id}`, {
                method: 'DELETE',
            });

            if (response.ok) {
                this.todos = this.todos.filter(t => t.id !== id);
                this.renderTodos();
            } else {
                throw new Error('Failed to delete todo');
            }
        } catch (error) {
            console.error('Failed to delete todo:', error);
            this.showError('Failed to delete todo');
        }
    }

    renderTodos() {
        const todosList = document.getElementById('todosList');

        if (this.todos.length === 0) {
            todosList.innerHTML = '<p class="loading">No todos yet. Add one above! 🚀</p>';
            return;
        }

        const todosHTML = this.todos.map(todo => {
            const createdAt = new Date(todo.created_at).toLocaleDateString();
            return `
                <div class="todo-item ${todo.completed ? 'completed' : ''}">
                    <div class="todo-header">
                        <div class="todo-title">${this.escapeHtml(todo.title)}</div>
                        <div class="todo-actions">
                            <button class="btn-complete" onclick="todoApp.toggleTodo(${todo.id})">
                                ${todo.completed ? '↩️ Undo' : '✅ Done'}
                            </button>
                            <button class="btn-delete" onclick="todoApp.deleteTodo(${todo.id})">
                                🗑️ Delete
                            </button>
                        </div>
                    </div>
                    ${todo.description ? `<div class="todo-description">${this.escapeHtml(todo.description)}</div>` : ''}
                    <div class="todo-meta">Created: ${createdAt}</div>
                </div>
            `;
        }).join('');

        todosList.innerHTML = todosHTML;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showError(message) {
        const todosList = document.getElementById('todosList');
        todosList.innerHTML = `<div class="error">${message}</div>`;
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.todoApp = new TodoApp();
});

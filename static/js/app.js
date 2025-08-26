// ASM Tracker - Main JavaScript

document.addEventListener('DOMContentLoaded', function() {
    // Auto-hide flash messages after 5 seconds
    const alerts = document.querySelectorAll('.alert[data-auto-hide="true"]');
    alerts.forEach(alert => {
        setTimeout(() => {
            alert.style.animation = 'fadeOut 0.3s ease-out';
            setTimeout(() => alert.remove(), 300);
        }, 5000);
    });

    // Add loading state to form submissions
    const forms = document.querySelectorAll('form[data-loading="true"]');
    forms.forEach(form => {
        form.addEventListener('submit', function() {
            const submitBtn = form.querySelector('button[type="submit"]');
            if (submitBtn) {
                submitBtn.classList.add('loading');
                submitBtn.disabled = true;
            }
        });
    });

    // Confirm delete actions
    const deleteLinks = document.querySelectorAll('a[data-confirm], button[data-confirm]');
    deleteLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            const message = this.getAttribute('data-confirm');
            if (!confirm(message)) {
                e.preventDefault();
                return false;
            }
        });
    });

    // Auto-focus first input field with error
    const errorInput = document.querySelector('.form-control.is-invalid');
    if (errorInput) {
        errorInput.focus();
    }
});

// Utility functions
window.ASMTracker = {
    // Show toast notification
    showToast: function(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `alert alert-${type} position-fixed top-0 end-0 m-3`;
        toast.style.zIndex = '9999';
        toast.innerHTML = `
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            ${message}
        `;
        document.body.appendChild(toast);
        
        // Auto-remove after 3 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.remove();
            }
        }, 3000);
    },

    // Format dates
    formatDate: function(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    },

    // Debounce function for search inputs
    debounce: function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

// Add fadeOut animation for CSS
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeOut {
        from { opacity: 1; }
        to { opacity: 0; }
    }
`;
document.head.appendChild(style);
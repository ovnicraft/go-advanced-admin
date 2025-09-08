// Go Advanced Admin JavaScript
// Enhanced functionality for search, delete modals, form controls, and bulk operations

// Global variables
let searchTimeout;

// Initialize all functionality when DOM is ready
$(document).ready(function() {
    console.log('Go Advanced Admin JavaScript initialized');
    
    // Initialize enhanced form controls
    initializeFormControls();
    
    // Initialize Go Advanced Admin functionality
    setupDeleteModals();
    setupSearch();
    setupBulkOperations();
});

// Initialize enhanced form controls (Select2, Flatpickr)
function initializeFormControls() {
    // Initialize Select2 dropdowns
    if (typeof $.fn.select2 !== 'undefined') {
        $('.select2').select2({
            theme: 'bootstrap-5'
        });
    }
    
    // Initialize Flatpickr date pickers
    if (typeof flatpickr !== 'undefined') {
        $('.flatpickr').flatpickr({
            dateFormat: "Y-m-d",
            allowInput: true
        });
        
        // DateTime picker
        $('[data-enable-time="true"]').flatpickr({
            enableTime: true,
            dateFormat: "Y-m-d H:i",
            allowInput: true,
            time_24hr: true
        });
    }
    
    // Initialize Select2 AJAX dropdowns
    $('[data-role="select2-ajax"]').each(function() {
        $(this).select2({
            theme: 'bootstrap-5',
            minimumInputLength: 1,
            ajax: {
                url: $(this).data("url"),
                dataType: 'json',
                data: function (params) {
                    return {
                        name: $(this).attr("name"),
                        term: params.term,
                    };
                }
            }
        });

        // Load existing data
        const existingData = $(this).data("json") || [];
        for (let i = 0; i < existingData.length; i++) {
            const data = existingData[i];
            const option = new Option(data.text, data.id, true, true);
            $(this).append(option).trigger('change');
        }
    });

    // Initialize Select2 tags
    $('[data-role="select2-tags"]').each(function() {
        $(this).select2({
            theme: 'bootstrap-5',
            tags: true,
            multiple: true,
        });

        const existingData = $(this).data("json") || [];
        for (let i = 0; i < existingData.length; i++) {
            const option = new Option(existingData[i], existingData[i], true, true);
            $(this).append(option).trigger('change');
        }
    });
}

// Setup delete confirmation modals
function setupDeleteModals() {
    // Handle delete button clicks with Bootstrap modal
    $(document).on('click', '[data-bs-target="#deleteModal"]', function(e) {
        e.preventDefault();
        
        const button = $(this);
        const deleteUrl = button.data('delete-url');
        const itemName = button.data('item-name') || 'this item';
        
        // Update modal content
        const modal = $('#deleteModal');
        const confirmButton = modal.find('#confirmDelete');
        
        // Update modal text
        modal.find('.modal-title').text('Are you sure?');
        modal.find('.modal-body div:last-child').text(`This will permanently delete ${itemName}. This action cannot be undone.`);
        
        // Set up confirm button click handler
        confirmButton.off('click').on('click', function() {
            performDelete(deleteUrl, modal);
        });
    });
}

// Perform AJAX delete operation
function performDelete(url, modal) {
    $.ajax({
        url: url,
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        success: function(data) {
            if (data.success) {
                // Close modal
                const bsModal = bootstrap.Modal.getInstance(modal[0]);
                if (bsModal) {
                    bsModal.hide();
                }
                
                // Show success message briefly, then reload
                if (data.message) {
                    showNotification(data.message, 'success');
                }
                
                // Reload page after short delay
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            } else {
                const errors = data.errors ? data.errors.join(', ') : 'Unknown error occurred';
                showNotification('Error: ' + errors, 'error');
            }
        },
        error: function(xhr, status, error) {
            console.error('Delete error:', error);
            let errorMessage = 'Failed to delete item';
            
            try {
                const response = JSON.parse(xhr.responseText);
                if (response.errors) {
                    errorMessage = response.errors.join(', ');
                } else if (response.message) {
                    errorMessage = response.message;
                }
            } catch (e) {
                errorMessage = xhr.responseText || errorMessage;
            }
            
            showNotification(errorMessage, 'error');
        }
    });
}

// Setup enhanced search functionality
function setupSearch() {
    const searchInput = $('#search-input');
    if (searchInput.length === 0) return;
    
    // Handle input with debouncing
    searchInput.on('input', function() {
        clearTimeout(searchTimeout);
        const query = $(this).val();
        
        if (query.length === 0) {
            // Clear search immediately if input is empty
            performSearch('');
            return;
        }
        
        searchTimeout = setTimeout(() => {
            performSearch(query);
        }, 500); // 500ms debounce
    });
    
    // Handle enter key
    searchInput.on('keypress', function(e) {
        if (e.which === 13) {
            e.preventDefault();
            clearTimeout(searchTimeout);
            performSearch($(this).val());
        }
    });
    
    // Handle search button click
    $(document).on('click', '[data-search]', function(e) {
        e.preventDefault();
        clearTimeout(searchTimeout);
        performSearch(searchInput.val());
    });
}

// Perform search operation
function performSearch(query) {
    const url = new URL(window.location);
    
    if (query && query.trim() !== '') {
        url.searchParams.set('q', query.trim());
    } else {
        url.searchParams.delete('q');
    }
    
    // Only reload if URL actually changed
    if (url.toString() !== window.location.toString()) {
        window.location.href = url.toString();
    }
}

// Setup bulk operations functionality
function setupBulkOperations() {
    const selectAll = $('#select-all');
    const rowCheckboxes = $('.row-checkbox');
    const bulkActionBar = $('#bulk-action-bar');
    const selectedCount = $('#selected-count');
    
    if (selectAll.length === 0 || bulkActionBar.length === 0) return;
    
    // Select all functionality
    selectAll.on('change', function() {
        const isChecked = this.checked;
        rowCheckboxes.prop('checked', isChecked);
        updateBulkActionBar();
    });
    
    // Individual checkbox handling
    rowCheckboxes.on('change', function() {
        updateBulkActionBar();
        
        // Update select-all checkbox state
        const checkedCount = rowCheckboxes.filter(':checked').length;
        const totalCount = rowCheckboxes.length;
        
        selectAll.prop('indeterminate', checkedCount > 0 && checkedCount < totalCount);
        selectAll.prop('checked', checkedCount === totalCount);
    });
    
    function updateBulkActionBar() {
        const checkedBoxes = rowCheckboxes.filter(':checked');
        const count = checkedBoxes.length;
        
        if (count > 0) {
            bulkActionBar.show();
            selectedCount.text(`${count} item${count > 1 ? 's' : ''} selected`);
        } else {
            bulkActionBar.hide();
        }
    }
}

// Bulk delete operation
function bulkDelete() {
    const checkedBoxes = $('.row-checkbox:checked');
    const ids = checkedBoxes.map(function() {
        return $(this).val() || $(this).data('id');
    }).get();
    
    if (ids.length === 0) {
        showNotification('No items selected', 'warning');
        return;
    }
    
    if (!confirm(`Delete ${ids.length} selected items? This action cannot be undone.`)) {
        return;
    }
    
    const currentPath = window.location.pathname;
    
    $.ajax({
        url: currentPath + '/bulk-delete',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        data: JSON.stringify({ ids: ids }),
        success: function(data) {
            if (data.success) {
                showNotification(data.message || `${data.data.deleted} items deleted successfully`, 'success');
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            } else {
                const errors = data.errors ? data.errors.join('\n') : 'Unknown error occurred';
                showNotification('Error: ' + errors, 'error');
            }
        },
        error: function(xhr, status, error) {
            console.error('Bulk delete error:', error);
            let errorMessage = 'Failed to delete items';
            
            try {
                const response = JSON.parse(xhr.responseText);
                if (response.errors) {
                    errorMessage = response.errors.join('\n');
                } else if (response.message) {
                    errorMessage = response.message;
                }
            } catch (e) {
                errorMessage = xhr.responseText || errorMessage;
            }
            
            showNotification(errorMessage, 'error');
        }
    });
}

// Clear selection
function clearSelection() {
    $('.row-checkbox').prop('checked', false);
    $('#select-all').prop('checked', false).prop('indeterminate', false);
    $('#bulk-action-bar').hide();
}

// Utility function to show notifications
function showNotification(message, type) {
    const alertClass = {
        'success': 'alert-success',
        'error': 'alert-danger',
        'warning': 'alert-warning',
        'info': 'alert-info'
    }[type] || 'alert-info';
    
    const alert = $(`
        <div class="alert ${alertClass} alert-dismissible fade show position-fixed" 
             style="top: 20px; right: 20px; z-index: 9999; min-width: 300px;">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `);
    
    $('body').append(alert);
    
    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        alert.alert('close');
    }, 5000);
}

// Global functions for template compatibility
window.bulkDelete = bulkDelete;
window.clearSelection = clearSelection;
window.performSearch = performSearch;

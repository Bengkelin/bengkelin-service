package models

// OrderStatus represents the status of an order
type OrderStatus uint

const (
	OrderStatusPending   OrderStatus = 0 // Order created, waiting for confirmation
	OrderStatusConfirmed OrderStatus = 1 // Bengkel confirmed the order
	OrderStatusInProgress OrderStatus = 2 // Service is being performed
	OrderStatusCompleted OrderStatus = 3 // Service completed successfully
	OrderStatusCancelled OrderStatus = 4 // Order cancelled by user or mitra
)

// String returns the string representation of OrderStatus
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "pending"
	case OrderStatusConfirmed:
		return "confirmed"
	case OrderStatusInProgress:
		return "in_progress"
	case OrderStatusCompleted:
		return "completed"
	case OrderStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// IsValid checks if the status is valid
func (s OrderStatus) IsValid() bool {
	return s >= OrderStatusPending && s <= OrderStatusCancelled
}

// CanTransitionTo checks if transition to new status is allowed
func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
	switch s {
	case OrderStatusPending:
		// From pending: can confirm, cancel
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		// From confirmed: can start work, cancel
		return newStatus == OrderStatusInProgress || newStatus == OrderStatusCancelled
	case OrderStatusInProgress:
		// From in progress: can complete, cancel (rare)
		return newStatus == OrderStatusCompleted || newStatus == OrderStatusCancelled
	case OrderStatusCompleted:
		// Completed orders cannot change status
		return false
	case OrderStatusCancelled:
		// Cancelled orders cannot change status
		return false
	default:
		return false
	}
}

// CancellationReason represents why an order was cancelled
type CancellationReason string

const (
	CancelledByUser          CancellationReason = "cancelled_by_user"
	CancelledByMitra         CancellationReason = "cancelled_by_mitra"
	CancelledBySystem        CancellationReason = "cancelled_by_system"
	CancelledNoShow          CancellationReason = "cancelled_no_show"
	CancelledPaymentFailed   CancellationReason = "cancelled_payment_failed"
	CancelledServiceUnavailable CancellationReason = "cancelled_service_unavailable"
)

// String returns the string representation of CancellationReason
func (r CancellationReason) String() string {
	return string(r)
}
package domain

type Product struct {
	ID       string
	MillID   string
	Type     ProductType
	Quantity float64
}

type ProductType string

const (
	ProductCPO    ProductType = "CPO"
	ProductKernel ProductType = "KERNEL"
)

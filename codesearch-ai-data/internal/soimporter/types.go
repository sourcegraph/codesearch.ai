package soimporter

type SOPostRow struct {
	ID               int     `xml:"Id,attr"`
	PostTypeID       uint8   `xml:"PostTypeId,attr"`
	ParentID         *int    `xml:"ParentId,attr"`
	AcceptedAnswerID *int    `xml:"AcceptedAnswerId,attr"`
	Title            *string `xml:"Title,attr"`
	Body             string  `xml:"Body,attr"`
	Score            int     `xml:"Score,attr"`
	Tags             *string `xml:"Tags,attr"`
	AnswerCount      *int    `xml:"AnswerCount,attr"`
	CreationDate     string  `xml:"CreationDate,attr"`
	LastEditDate     string  `xml:"LastEditDate,attr"`
}

type SOQuestion struct {
	ID               int
	Title            string
	Tags             string
	Score            int
	AcceptedAnswerID *int
	CreationDate     string
	LastEditDate     string
}

type SOAnswer struct {
	ID           int
	Body         string
	Score        int
	ParentID     int
	CreationDate string
	LastEditDate string
}

CREATE TABLE decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    max_user_id BIGINT,
    type TEXT NOT NULL DEFAULT 'loan',
    status TEXT NOT NULL DEFAULT 'calculated',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE decision_inputs (
    decision_id UUID PRIMARY KEY REFERENCES decisions(id) ON DELETE CASCADE,
    loan_amount NUMERIC(14,2) NOT NULL,
    loan_term INT NOT NULL,
    interest_rate NUMERIC(5,2) NOT NULL,
    revenue NUMERIC(14,2) NOT NULL,
    expenses NUMERIC(14,2) NOT NULL,
    reserve NUMERIC(14,2) NOT NULL,
    revenue_growth NUMERIC(6,2) NOT NULL,
    expense_growth NUMERIC(6,2) NOT NULL,
    purpose TEXT NOT NULL DEFAULT ''
);

CREATE TABLE scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    revenue NUMERIC(14,2) NOT NULL,
    expenses NUMERIC(14,2) NOT NULL,
    loan_payment NUMERIC(14,2) NOT NULL,
    cash_flow NUMERIC(14,2) NOT NULL,
    debt_load_pct NUMERIC(6,2) NOT NULL,
    reserve_after_3m NUMERIC(14,2) NOT NULL,
    reserve_months INT,
    consequences JSONB NOT NULL DEFAULT '[]'
);

CREATE INDEX idx_decisions_user_created ON decisions (max_user_id, created_at DESC);
CREATE INDEX idx_scenarios_decision ON scenarios (decision_id);

The goal is to build a "Smart Alert" system that tells the user how much to invest based on two factors:

Value Averaging: Keeping the portfolio on a steady growth path.
Market Valuation: Buying more when Nifty PE is low, and less when it is high.
Proposed Changes
[New] Automated Alert Service
Architecture:
A long-running Go process or a binary triggered by a Cron job.
Scheduler: Use robfig/cron to run the analysis every morning (e.g., 9:15 AM).
Core Engine: The "Cost-Optimized Drawdown" Model
For each asset i:
DD_i = (ATH_i - Current_i) / ATH_i (Market Discount).
L_i = (Current_i - BuyPrice_i) / BuyPrice_i (Portfolio Profit/Loss).
Decision Logic:
If L_i > 0 (Position is in Profit):
Strategy: "Don't Average Up".
Action: Invest only the Base/Minimum amount (e.g., ₹2k), regardless of the drawdown. This preserves your low cost-basis.
If L_i <= 0 (Position is at cost or in loss):
Strategy: "Aggressive Averaging".
Action: Use the full Multiplier logic based on DD_i (Drawdown).
Exception (Indices): For Nifty 50, we ignore buy-price and always follow Drawdown/PE, as we want to accumulate index units perpetually.
Rule: The "Tactical Bulk Buy" (Nifty PE 18)
Same as before: Overrides all constraints for Nifty 50.
Notification Layer:
Include a "Psychology Note" in the alert: "Skipping aggressive allocation for X because it's still above your average cost."
Persistence:
config.json stores:
Telegram Bot Token & Chat ID.
MonthlyBudget: 35000
Assets:
{"Symbol": "NIFTY_50", "Type": "Index", "Target": 10000}
{"Symbol": "PPFAS", "SchemeCode": "122639", "Type": "MF", "Target": 10000}
{"Symbol": "WhiteOak", "SchemeCode": "148858", "Type": "MF", "Target": 5000}
{"Symbol": "Helios", "SchemeCode": "150774", "Type": "MF", "Target": 5000}
{"Symbol": "Debt", "Type": "Cash", "Target": 5000}
Workflow:
Wake up at 9:15 AM.
Fetch data for all 10+ assets.
Analyze drawdowns and PE.
If any asset is in "Buy Zone" (>X% drawdown) or PE < 18, fire the alert.
Verification Plan
Automated Tests
Test the scheduler and the message formatting logic.
Mock the data fetching to verify "Dip Alerts" are only sent when thresholds are met.
Manual Verification
Run the service: go run main.go
Verify Telegram: Ensure a message is received on the configured chat ID.
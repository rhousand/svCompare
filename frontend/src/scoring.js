/**
 * Scoring sections and questions — frontend copy of the canonical list.
 * Must stay in sync with backend/internal/scoring/scoring.go.
 *
 * Tooltip text is derived from sailboat_buyers_guide.md.
 */
export const SECTIONS = [
  {
    name: 'Ownership & History',
    weight: 0.20,
    questions: [
      {
        id: 1,
        text: 'How many owners has the boat had?',
        tooltip: 'Good: 1–2 owners. Fewer owners generally means more consistent maintenance culture.\nAvoid: 3+ owners in a short period, or a history of quick resales — often signals recurring problems or disappointment.',
      },
      {
        id: 2,
        text: 'Has it ever been in a charter fleet? If so, for how many seasons?',
        tooltip: 'Good: Never chartered, or private use only.\nCaution: 1–2 charter seasons with documented professional maintenance isn\'t a dealbreaker — charter companies often service boats on a strict schedule.\nAvoid: 3+ charter seasons, especially with a bareboat operation. Expect worn upholstery, stressed rigging, abused engines, and deferred cosmetic repairs.',
      },
      {
        id: 3,
        text: 'Is there a complete maintenance log?',
        tooltip: 'Good: Logbook or digital record showing regular oil changes, impeller replacements, rigging checks, haulouts, etc. This signals an owner who cared.\nAvoid: "I just did stuff as needed" with nothing to show for it. No records = no accountability. Budget for a full systems audit.',
      },
      {
        id: 4,
        text: 'Has the boat ever run aground or been in a collision? Any insurance claims?',
        tooltip: 'Good: Clean history with no claims.\nCaution: A soft grounding in sand with documented inspection is not necessarily serious.\nAvoid: Any collision or hard grounding with no subsequent survey. Keel damage, rudder damage, and hull stress fractures can be hidden cosmetically.',
      },
    ],
  },
  {
    name: 'Engine & Mechanical',
    weight: 0.20,
    questions: [
      {
        id: 5,
        text: 'What are the current engine hours?',
        tooltip: 'Good: Under 1,000 hours for a boat 5–7 years old (~100–150 hrs/year for light use). Well-maintained diesels can go 3,000–5,000+ hours but cost more to maintain as they age.\nCaution: 1,500–2,500 hours — not a dealbreaker but factor in upcoming service costs.\nAvoid: Very low hours (under 200) on a boat that is several years old. Infrequently run diesels develop problems from sitting.',
      },
      {
        id: 6,
        text: 'When was the last engine service, and what was done?',
        tooltip: 'Good: Full service within the last 12 months or 100 hours — oil, filters, impeller, belts, zincs.\nAvoid: "It runs fine" with no documented recent service. Budget $500–$1,500 for a full diesel service immediately if records are absent.',
      },
      {
        id: 7,
        text: 'Is the diesel impeller, belts, and heat exchanger current?',
        tooltip: 'Good: Impeller replaced within the last season. Belt and heat exchanger inspected.\nAvoid: Unknown impeller history — a failed impeller destroys the heat exchanger and can cause the engine to overheat in minutes. This is a cheap part (~$30) with expensive consequences if neglected.',
      },
    ],
  },
  {
    name: 'Sails & Rig',
    weight: 0.15,
    questions: [
      {
        id: 9,
        text: 'What year are the sails from? Are they original?',
        tooltip: 'Good: Sails replaced or lightly used — original sails with fewer than 500 hours still have life. Sails from a previous owner\'s upgrade are a bonus.\nCaution: Original sails are fine if the boat hasn\'t been sailed hard. UV damage and delamination are the killers — inspect in person.\nAvoid: Sails that are stiff, delaminating (layers separating), have heavy UV strip fading, or numerous patches. New offshore-quality sails for a 40-footer run $8,000–$20,000+.',
      },
      {
        id: 10,
        text: 'What sail inventory is included?',
        tooltip: 'Good: Full main, furling headsail, and a bonus sail (code zero, asymmetric spinnaker, storm jib). More inventory = more value.\nAvoid: Just a main and furling jib with no extras. Fine for coastal sailing, but factor in what you\'d need to add for your intended use.',
      },
      {
        id: 11,
        text: 'Has the mast ever been unstepped? When? Any issues found?',
        tooltip: 'Good: Unstepped for a rigging replacement or inspection with no issues found.\nAvoid: Never unstepped on an older boat — wiring and rigging inside the mast degrades invisibly. Any vague answers about what was found during an unstepping deserves follow-up.',
      },
      {
        id: 15,
        text: 'What is the age and condition of the standing rigging?',
        tooltip: 'Ask this separately from other rigging questions — sellers sometimes give inconsistent answers.\nGood sign: Consistent answer with other rigging questions, showing the seller tracks maintenance carefully.\nRed flag: A different answer than given elsewhere — suggests the seller isn\'t maintaining or tracking the boat attentively.\nStanding rigging (shrouds, stays) should be replaced within 10 years or 10,000 nm. Budget $3,000–$8,000+ for a full re-rig if it\'s due.',
      },
      {
        id: 26,
        text: 'What is the condition of the running rigging? When was it last replaced?',
        tooltip: 'Good: Running rigging is soft and pliable, not stiff or fuzzy. Recent replacement noted in the maintenance log. Halyards, sheets, and control lines in good condition.\nAvoid: Stiff, glazed, or fuzzy lines — they\'re fatigued and prone to failure under load. Original running rigging on an older boat with no log of replacement. Budget $1,000–$3,000+ to replace running rigging on a 40-footer.',
      },
    ],
  },
  {
    name: 'Systems',
    weight: 0.15,
    questions: [
      {
        id: 12,
        text: 'What is the battery bank configuration (capacity, age, type)?',
        tooltip: 'Good: Lithium (LiFePO4) bank or AGM batteries less than 3 years old with 200+ Ah capacity. Proper battery monitor (Victron, etc.) installed.\nCaution: Older AGMs can appear to hold charge but fail under load. Ask to see a battery test.\nAvoid: Original batteries on a boat 5+ years old with no upgrades. Lead-acid batteries have a 3–5 year lifespan. Budget $1,500–$6,000+ for a proper replacement bank.',
      },
      {
        id: 13,
        text: 'Is there shore power? What is the inverter/charger setup?',
        tooltip: 'Good: Working shore power inlet, GFCI-protected outlets, and a quality charger (Victron, Mastervolt). Inverter is a bonus for offshore use.\nAvoid: Corroded shore power connections or any DIY electrical work. Boat electrical fires are a leading cause of total losses. If the wiring looks messy, budget for a full inspection by a marine electrician.',
      },
      {
        id: 14,
        text: 'Is the head a holding tank system? Is it certified compliant?',
        tooltip: 'Good: Y-valve system with a properly sized holding tank and macerator pump. Legally compliant for no-discharge zones.\nAvoid: Direct overboard discharge without a holding tank — illegal in most US waters. Retrofitting a compliant system costs $500–$2,000 and is disruptive.',
      },
    ],
  },
  {
    name: 'Survey & Hull Condition',
    weight: 0.20,
    questions: [
      {
        id: 16,
        text: 'Has a recent marine survey been conducted? Can I see it?',
        tooltip: 'Good: Survey within the last 2 years available for review.\nImportant: The seller\'s survey protects the seller. Always commission your own independent survey — typically $20–$25 per foot.\nAvoid: No survey, or any resistance to letting you conduct one. No reputable seller blocks an independent survey. Walk away if they do.',
      },
      {
        id: 17,
        text: 'When was the boat last hauled? What bottom paint is on it?',
        tooltip: 'Good: Hauled within the last 12–18 months. Fresh antifouling paint. Thru-hulls and zincs inspected.\nAvoid: Last hauled more than 2 years ago, or unknown. Expect fouled running gear, degraded zinc protection, and potentially barnacled thru-hulls. Budget $1,500–$3,000 for a haulout + bottom job.',
      },
      {
        id: 18,
        text: 'Is there any osmotic blistering or delamination on the hull?',
        tooltip: 'Good: Clean hull with no blistering, or minor surface blisters already treated.\nCaution: Small surface blisters are cosmetic and treatable. Get details confirmed in the survey.\nAvoid: Deep structural blisters or delamination. A full osmotic treatment (barrier coat) can run $5,000–$15,000+ depending on severity and boat size.',
      },
      {
        id: 19,
        text: 'What is the condition of the rudder bearings and keel bolts?',
        tooltip: 'Good: Rudder is firm with no slop when wiggled. Keel bolts inspected at last haulout with no weeping rust stains.\nAvoid: Any slop in the rudder — it will worsen and eventually fail. Rust staining around keel bolts is a serious red flag requiring immediate expert evaluation. Keel re-bedding is a major, expensive project.',
      },
    ],
  },
  {
    name: 'Electronics & Safety',
    weight: 0.10,
    questions: [
      {
        id: 20,
        text: 'What electronics are included (chartplotter, AIS, VHF, autopilot)?',
        tooltip: 'Good: Modern chartplotter (Garmin, Raymarine, B&G) with current charts, AIS transponder (not just receiver), working VHF, autopilot, and wind/speed/depth instruments. Radar is a strong bonus.\nCaution: Older electronics work fine but may lack integration. Factor in upgrade costs to your offer.\nAvoid: Non-functional electronics the seller "hasn\'t gotten around to fixing." A full electronics suite for a 40-footer runs $5,000–$20,000+.',
      },
      {
        id: 21,
        text: 'Is there a windlass? Watermaker? Generator?',
        tooltip: 'Good: Electric windlass in working condition is near-essential for shorthanded sailing. A watermaker and generator are serious value-adds for offshore or liveaboard use.\nAvoid: A windlass that "works sometimes" — chain jams at the worst moments. Verify function in person. A new windlass runs $800–$2,500 installed.',
      },
      {
        id: 22,
        text: 'What safety gear is included and what are the expiration dates?',
        tooltip: 'Good: Life raft serviced within 3 years, flares current, EPIRBs registered and within hydrostatic test date, jacklines and tethers in good condition.\nAvoid: Expired safety gear doesn\'t protect you and costs real money to replace. A 4-person offshore life raft costs $3,000–$6,000+ to purchase or $400–$800 to service. Factor all expired gear into your negotiation.',
      },
    ],
  },
  {
    name: 'Transaction',
    weight: 0.00,
    questions: [
      {
        id: 23,
        text: 'Is the price negotiable? Has the listing price been reduced?',
        tooltip: 'Good: Seller acknowledges room to negotiate. A price reduction history shows motivation to sell.\nCaution: "Firm on price" isn\'t necessarily bad if the boat is priced fairly — verify against comparables first.\nStrategy: Always make a written offer 10–15% below asking after your survey. Use any survey findings as additional negotiating leverage.',
      },
      {
        id: 24,
        text: 'Where is the boat located, and what are the current slip costs?',
        tooltip: 'Good: Boat is in the water, accessible for a sea trial, in a reputable marina.\nCaution: "In storage" or "on the hard" means you can\'t do a full sea trial until it\'s launched — factor in that cost (~$500–$1,500).\nAvoid: Boats stored in unfamiliar or remote locations far from professional resources. Buying sight-unseen is high risk.',
      },
      {
        id: 25,
        text: 'Is it available for a sea trial prior to purchase?',
        tooltip: 'Good: Seller says yes without hesitation.\nAvoid: Any resistance to a sea trial. You must sail the boat before buying. A sea trial reveals engine performance under load, sail handling, instrument function, unusual sounds/smells, and how the boat feels. It is non-negotiable.',
      },
    ],
  },
]

/**
 * Build a lookup map from a boat's scores array.
 * @param {Array} scores - Array of { question_id, value } objects
 * @returns {Object} Map of questionId -> value
 */
export function buildScoreMap(scores) {
  const map = {}
  for (const s of scores || []) {
    map[s.question_id] = s.value
  }
  return map
}

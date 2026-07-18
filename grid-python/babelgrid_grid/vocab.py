"""The fixed vocabulary the mock STT draws from, and the phrase-table
dictionaries the MT layer translates with (en -> es/fr/de)."""

# Words the (mock) STT can emit. Tokens from the DSP index into this bank.
WORD_BANK = [
    "the", "quick", "meeting", "started", "late", "team", "shipped", "new",
    "build", "today", "please", "review", "pull", "request", "before", "noon",
    "we", "need", "more", "tests", "docs", "hello", "world", "thanks",
]

# en -> target dictionaries. Words not present fall through unchanged.
DICT = {
    "es": {
        "the": "el", "quick": "rápido", "meeting": "reunión", "started": "empezó",
        "late": "tarde", "team": "equipo", "shipped": "envió", "new": "nuevo",
        "build": "compilación", "today": "hoy", "please": "por favor",
        "review": "revisar", "pull": "extraer", "request": "solicitud",
        "before": "antes", "noon": "mediodía", "we": "nosotros", "need": "necesitamos",
        "more": "más", "tests": "pruebas", "docs": "documentos", "hello": "hola",
        "world": "mundo", "thanks": "gracias",
    },
    "fr": {
        "the": "le", "quick": "rapide", "meeting": "réunion", "started": "commencé",
        "late": "tard", "team": "équipe", "shipped": "livré", "new": "nouveau",
        "build": "build", "today": "aujourd'hui", "please": "s'il vous plaît",
        "review": "réviser", "pull": "tirer", "request": "demande",
        "before": "avant", "noon": "midi", "we": "nous", "need": "avons besoin",
        "more": "plus", "tests": "tests", "docs": "documents", "hello": "bonjour",
        "world": "monde", "thanks": "merci",
    },
    "de": {
        "the": "die", "quick": "schnelle", "meeting": "besprechung", "started": "begann",
        "late": "spät", "team": "team", "shipped": "lieferte", "new": "neue",
        "build": "build", "today": "heute", "please": "bitte",
        "review": "überprüfen", "pull": "ziehen", "request": "anfrage",
        "before": "vor", "noon": "mittag", "we": "wir", "need": "brauchen",
        "more": "mehr", "tests": "tests", "docs": "dokumente", "hello": "hallo",
        "world": "welt", "thanks": "danke",
    },
}

SUPPORTED = tuple(DICT.keys())

# Benchmark — Token measurement

Mesures réelles de tokens sur un corpus fixe de commandes représentatives,
en mode **human** et **agent**.

## Setup

```bash
uv venv bench/.venv
uv pip install tiktoken --python bench/.venv/bin/python
```

## Utilisation

```bash
# 1. Collecter les sorties brutes
bash bench/corpus.sh

# 2. Compter les tokens
bench/.venv/bin/python bench/count_tokens.py

# 3. Comparer human vs agent
bench/.venv/bin/python bench/compare.py

# Ou tout en une commande
just bench
```

## Fichiers

| Fichier | Rôle |
|---|---|
| `corpus.sh` | Exécute les 4 commandes de référence en mode human et agent |
| `count_tokens.py` | Compte les tokens (cl100k_base) sur chaque sortie brute |
| `compare.py` | Génère le rapport comparatif human vs agent |
| `check_regression.py` | Vérifie que les tokens agent n'ont pas augmenté |
| `results/raw/` | Sorties brutes JSON par mode/commande |
| `results/tokens.csv` | Résultats de comptage |
| `results/report.md` | Rapport comparatif (généré par compare.py) |
| `results/baseline.csv` | Baseline de référence pour la régression |

## Encodage

`cl100k_base` (tiktoken) — compatible GPT-4 et bonne approximation Claude.

# nitfeed

Industrial-grade anti-doomscroll technology:

1. Write a list of Twitter handles in `accounts.txt`.
2. Run `main.go`.
3. Open your new file (`output.html`).

`output.html` will contain a sorted Twitter feed with recent tweets from the accounts in `accounts.txt`.

## Example

Example `accounts.txt`:

```
Teknium1
10x_er
ctjlewis
xlr8harder
ID_AA_Carmack
goth600
realGeorgeHotz
Soul0Engineer
tekbog
yacineMTB
tszzl
cto_junior
main_horse
teortaxesTex
aicrumb
cloud11665
AuroraNemoia
4evaBehindSOTA
ylecun
arthurmensch
MistralAI
ClementDelangue
AIatMeta
JeffDean
ggerganov
jeremyphoward
```

Run with:

```bash
go run main.go
```
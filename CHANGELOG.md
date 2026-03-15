# CHANGELOG

모든 Git 커밋 이력을 최신순으로 기록합니다. 새 커밋은 표 최상단에 추가합니다.

| 일시 | 유형 | 범위 | 변경내용 (목적 포함) |
|---|---|---|---|
| 2026-03-14 21:00 | chore | gitignore | notion.exe 바이너리 gitignore 추가 — Windows 빌드 결과물 추적 방지 |
| 2026-03-14 20:10 | fix | block | table_row children을 table{} 내부로 이동 — Notion API 스펙 준수 ('table.children should be defined' 오류 수정) |
| 2026-03-14 19:52 | feat | block | GFM 테이블 파싱 + 인라인 서식(bold/italic/code/link/strike) 지원 추가 — 노션 CLI로 마크다운 표 업로드 시 깨지던 문제 근본 해결 |

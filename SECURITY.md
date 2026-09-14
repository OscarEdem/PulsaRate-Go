# Security Policy
<!-- -------------------------------------------------------------------------------------------------------------------                                                                                                                                                                               #*eddiere -->

## Supported Versions

The following versions of **PulsaRate-Go** currently receive security updates:

| Version | Supported          |
| :------ | :----------------- |
| `main` (latest) | ✅ Actively supported |
| Older tags      | ❌ No longer supported — please upgrade |

---

## Reporting a Vulnerability

> [!CAUTION]
> **DO NOT open a public GitHub Issue for security vulnerabilities.** Public disclosure before a fix is available puts all users at risk.

### Private Disclosure Process

1. **Email:** Send a detailed report to the maintainer via GitHub's private vulnerability reporting feature:
   👉 [https://github.com/aeroforge-labs/PulsaRate-Go/security/advisories/new](https://github.com/aeroforge-labs/PulsaRate-Go/security/advisories/new)

2. **What to include in your report:**
   - A clear description of the vulnerability
   - The component affected (e.g., `pkg/limiter/`, `pkg/middleware/`, Redis Lua scripts)
   - Steps to reproduce or a proof-of-concept
   - Potential impact assessment (e.g., rate limit bypass, denial of service, data exposure)
   - Your suggested fix or mitigation, if any

3. **Response Timeline:**
   | Milestone | Target |
   | :--- | :--- |
   | Acknowledgement of report | ≤ 48 hours |
   | Initial triage & severity assessment | ≤ 5 business days |
   | Patch development & internal testing | ≤ 14 business days (severity-dependent) |
   | Public advisory & patched release | Coordinated with reporter |

---

## Security Threat Model

PulsaRate-Go is a **rate-limiting library** embedded in production Go services. The following threat categories are in-scope for security reports:

| Threat | Description |
| :--- | :--- |
| **Rate Limit Bypass** | Logic bugs allowing a client to exceed configured limits without triggering a reject |
| **Race Conditions** | Data races in `sync/atomic` operations or Redis lease management that affect correctness |
| **Redis Injection** | Malformed inputs that could affect the Lua scripts executed against Redis |
| **Denial of Service** | Resource exhaustion bugs (goroutine leak, memory leak) introduced in PulsaRate itself |
| **Unsafe Defaults** | Configuration defaults that are insecure or misleading for production deployments |

The following are **out of scope**:
- Vulnerabilities in third-party dependencies (report those upstream)
- Redis server misconfigurations in the user's infrastructure
- Issues in user application code that misuse the PulsaRate API

---

## Disclosure Policy

We follow **Coordinated Vulnerability Disclosure (CVD)**:

1. Reporter submits a private advisory.
2. Maintainer acknowledges, triages, and develops a fix in a private branch.
3. A patched release is published.
4. A public GitHub Security Advisory is issued crediting the reporter (unless anonymity is requested).

---

## Preferred Languages

We prefer all security communications in **English**.

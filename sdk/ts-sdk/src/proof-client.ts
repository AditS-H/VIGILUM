// ProofVerificationClient.ts - TypeScript client for proof verification API

/**
 * Request/Response types for proof verification API
 */

export interface GenerateChallengeRequest {
  user_id: string;
  verifier_address: string;
}

export interface GenerateChallengeResponse {
  challenge_id: string;
  issued_at: string;
  expires_at: string;
  ttl_seconds: number;
}

export interface SubmitProofRequest {
  challenge_id: string;
  proof_data: string; // hex-encoded
  timing_variance: number;
  gas_variance: number;
  proof_nonce: string;
}

export interface SubmitProofResponse {
  is_valid: boolean;
  verification_score: number;
  verification_result: string;
  risk_score_reduction: number;
  proof_id?: string;
  verified_at?: string;
  message: string;
}

export interface GetUserProofsResponse {
  user_id: string;
  proof_count: number;
  proofs: ProofInfo[];
  average_score: number;
  page_info: PaginationInfo;
}

export interface ProofInfo {
  id: string;
  proof_hash: string;
  verification_score: number;
  verified_at?: string;
  expires_at: string;
  created_at: string;
  verifier_address: string;
}

export interface GetVerificationScoreResponse {
  user_id: string;
  verification_score: number;
  proof_count: number;
  verified_proof_count: number;
  is_verified: boolean;
  last_verified_at?: string;
  risk_score: number;
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
  status_code: number;
  timestamp: string;
}

/**
 * ProofVerificationClient provides methods to interact with proof verification API
 */
export interface ProofClientOptions {
  /** Base URL for the API (omit trailing slash). Defaults to `/api/v1`. */
  baseUrl?: string;
  /** Optional fetch implementation for testing. Defaults to global fetch. */
  fetcher?: typeof fetch;
}

export class ProofVerificationClient {
  private readonly baseUrl: string;
  private readonly fetcher: typeof fetch;

  constructor(options: ProofClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? '/api/v1';
    this.fetcher = options.fetcher ?? fetch;
  }

  /**
   * Generate a proof challenge for a user
   */
  async generateChallenge(
    userId: string,
    verifierAddress: string
  ): Promise<GenerateChallengeResponse> {
    const request: GenerateChallengeRequest = {
      user_id: userId,
      verifier_address: verifierAddress,
    };

    return this.post<GenerateChallengeResponse>('/proofs/challenges', request);
  }

  /**
   * Submit a proof for verification
   */
  async submitProof(
    challengeId: string,
    proofData: Uint8Array,
    timingVariance: number,
    gasVariance: number,
    proofNonce: string
  ): Promise<SubmitProofResponse> {
    const request: SubmitProofRequest = {
      challenge_id: challengeId,
      proof_data: this.bytesToHex(proofData),
      timing_variance: timingVariance,
      gas_variance: gasVariance,
      proof_nonce: proofNonce,
    };

    return this.post<SubmitProofResponse>('/proofs/verify', request);
  }

  /**
   * Get user's proofs with pagination
   */
  async getUserProofs(
    userId: string,
    page: number = 1,
    limit: number = 10
  ): Promise<GetUserProofsResponse> {
    const params = new URLSearchParams({
      user_id: userId,
      page: page.toString(),
      limit: limit.toString(),
    });

    return this.get<GetUserProofsResponse>(`/proofs?${params.toString()}`);
  }

  /**
   * Get user's verification score
   */
  async getVerificationScore(userId: string): Promise<GetVerificationScoreResponse> {
    const params = new URLSearchParams({
      user_id: userId,
    });

    return this.get<GetVerificationScoreResponse>(`/verification-score?${params.toString()}`);
  }

  /**
   * Get challenge status
   */
  async getChallengeStatus(challengeId: string): Promise<any> {
    return this.get(`/proofs/challenges/${challengeId}`);
  }

  /**
   * Check service health
   */
  async getHealth(): Promise<any> {
    return this.get(`/health`);
  }

  /**
   * Verify firewall proof (alias endpoint)
   */
  async verifyFirewallProof(request: SubmitProofRequest): Promise<SubmitProofResponse> {
    return this.post<SubmitProofResponse>(`/firewall/verify-proof`, request);
  }

  // ---------------------------------------------------------------------------
  // Internal HTTP helpers
  // ---------------------------------------------------------------------------

  private async get<T>(path: string): Promise<T> {
    const res = await this.fetcher(`${this.baseUrl}${path}`, {
      method: 'GET',
      headers: { 'Accept': 'application/json' },
    });

    if (!res.ok) {
      throw await this.toError(res);
    }

    return (await res.json()) as T;
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const res = await this.fetcher(`${this.baseUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      throw await this.toError(res);
    }

    return (await res.json()) as T;
  }

  private async toError(res: Response): Promise<Error> {
    let message = `HTTP ${res.status}`;
    try {
      const data = await res.json();
      if (typeof data?.message === 'string') {
        message = data.message;
      }
    } catch (_) {
      // Ignore JSON parse errors
    }
    return new Error(message);
  }

  /**
   * Helper: Convert bytes to hex string
   */
  private bytesToHex(bytes: Uint8Array): string {
    return Array.from(bytes)
      .map(byte => byte.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * Helper: Convert hex string to bytes
   */
  static hexToBytes(hex: string): Uint8Array {
    const normalized = hex.startsWith('0x') ? hex.slice(2) : hex;
    if (normalized.length % 2 !== 0) {
      throw new Error('Hex string must have an even length');
    }

    const result = new Uint8Array(normalized.length / 2);
    for (let i = 0; i < normalized.length; i += 2) {
      const byte = normalized.substr(i, 2);
      const parsed = parseInt(byte, 16);
      if (Number.isNaN(parsed)) {
        throw new Error('Hex string contains invalid characters');
      }
      result[i / 2] = parsed;
    }
    return result;
  }
}

export interface ElectionResult {
    id: number;
    name: string;
    dob: string;
    bio_link: string;
    image_url: string;
    policy: string;
    vote_count: number;
    percentage: string;
    create_at: Date;
}

export interface ChartElection{
    name: string;
    value: number;
}
import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { FaUserCircle, FaQrcode, FaShareAlt, FaEdit, FaSignOutAlt, FaCalendarPlus, FaFileMedical, FaNotesMedical, FaHospital, FaChevronDown, FaChevronUp, FaDownload } from 'react-icons/fa';
import api from '../services/api';

const Profile = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(true);
    const [profile, setProfile] = useState(null);
    const [isEditing, setIsEditing] = useState(false);
    const [shareData, setShareData] = useState(null);
    const [error, setError] = useState('');

    const [formData, setFormData] = useState({
        full_name: '',
        date_of_birth: '',
        gender: 'Other',
        phone: '',
        address: ''
    });

    const [visits, setVisits] = useState([]);
    const [prescriptions, setPrescriptions] = useState([]);
    const [showAddVisit, setShowAddVisit] = useState(false);
    const [newVisit, setNewVisit] = useState({
        hospital_name: '',
        visit_date: new Date().toISOString().split('T')[0],
        reason: '',
        notes: ''
    });
    const [visitFile, setVisitFile] = useState(null);
    const [isSubmittingVisit, setIsSubmittingVisit] = useState(false);

    useEffect(() => {
        fetchProfile();
        fetchTimeline();
        fetchPrescriptions();
    }, []);

    const fetchProfile = async () => {
        try {
            const response = await api.get('/patients/profile');
            const profileData = response.data.data;
            setProfile(profileData);
            setFormData({
                full_name: profileData.full_name,
                date_of_birth: profileData.date_of_birth.split('T')[0],
                gender: profileData.gender,
                phone: profileData.phone,
                address: profileData.address
            });
            if (profileData.share_token) {
                setShareData({
                    share_token: profileData.share_token,
                    share_url: `/api/patients/share/${profileData.share_token}` // Just for local reference
                });
            }
        } catch (err) {
            if (err.response?.status === 404) {
                setProfile(null);
                setIsEditing(true); // Force create mode
            } else if (err.response?.status === 401) {
                navigate('/login');
            } else {
                setError('Failed to load profile');
            }
        } finally {
            setLoading(false);
        }
    };

    const fetchTimeline = async () => {
        try {
            const response = await api.get('/timeline/my');
            setVisits(response.data.data || []);
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to fetch timeline');
        }
    };

    const fetchPrescriptions = async () => {
        try {
            const response = await api.get('/prescriptions/my');
            setPrescriptions(response.data.data || []);
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to fetch prescriptions');
        }
    };

    const handleSave = async (e) => {
        e.preventDefault();
        try {
            if (profile) {
                await api.put('/patients/profile', formData);
            } else {
                await api.post('/patients/profile', formData);
            }
            setIsEditing(false);
            fetchProfile();
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to save profile');
        }
    };

    const handleAddVisit = async (e) => {
        e.preventDefault();
        setError(''); // Clear previous errors
        setIsSubmittingVisit(true);
        try {
            const visitResponse = await api.post('/timeline/visits', newVisit);

            if (visitResponse.data && visitResponse.data.data && visitResponse.data.data.id) {
                const visitId = visitResponse.data.data.id;

                if (visitFile) {
                    const formData = new FormData();
                    formData.append('file', visitFile);
                    formData.append('visit_id', visitId);
                    await api.post('/prescriptions/upload', formData, {
                        headers: { 'Content-Type': 'multipart/form-data' }
                    });
                }

                setShowAddVisit(false);
                setNewVisit({
                    hospital_name: '',
                    visit_date: new Date().toISOString().split('T')[0],
                    reason: '',
                    notes: ''
                });
                setVisitFile(null);
                fetchTimeline();
                fetchPrescriptions();
            } else {
                throw new Error('Invalid response from server');
            }
        } catch (err) {
            setError(err.response?.data?.error || err.message || 'Failed to add visit');
        } finally {
            setIsSubmittingVisit(false);
        }
    };

    const downloadPrescription = async (id, fileName) => {
        try {
            const response = await api.get(`/prescriptions/download?id=${id}`, {
                responseType: 'blob'
            });
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            link.remove();
        } catch (err) {
            setError('Failed to download prescription');
        }
    };
    const generateShareLink = async () => {
        try {
            const response = await api.post('/patients/share');
            setShareData(response.data.data);
        } catch (err) {
            setError('Failed to generate share link');
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        navigate('/login');
    };

    if (loading) return <div className="container" style={{ padding: '2rem', textAlign: 'center' }}>Loading...</div>;

    return (
        <div style={{ minHeight: '100vh', background: 'var(--bg-light)' }}>
            {/* Navbar */}
            <nav style={{ background: 'var(--white)', padding: '1rem 0', boxShadow: 'var(--shadow-sm)' }}>
                <div className="container" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <h1 style={{ color: 'var(--primary)', margin: 0, fontSize: '1.5rem', fontWeight: 700 }}>Health Bar</h1>
                    <button onClick={handleLogout} className="btn btn-outline" style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}>
                        <FaSignOutAlt style={{ marginRight: '0.5rem' }} /> Logout
                    </button>
                </div>
            </nav>

            <main className="container" style={{ padding: '2rem 1rem' }}>
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.5 }}
                >
                    {error && <div className="error-msg">{error}</div>}

                    {!profile && isEditing && (
                        <div className="card" style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <h2 className="title" style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Create Your Patient Profile</h2>
                            <ProfileForm
                                formData={formData}
                                setFormData={setFormData}
                                handleSave={handleSave}
                                isNew={true}
                            />
                        </div>
                    )}

                    {profile && !isEditing && (
                        <div style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <div className="card" style={{ marginBottom: '2rem' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
                                    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                                        <FaUserCircle size={60} color="var(--primary-light)" style={{ color: 'var(--primary)' }} />
                                        <div>
                                            <h2 style={{ fontSize: '1.5rem', fontWeight: 600 }}>{profile.full_name}</h2>
                                            <p style={{ color: 'var(--text-muted)' }}>Patient ID: {profile.id.substring(0, 8)}...</p>
                                        </div>
                                    </div>
                                    <button onClick={() => setIsEditing(true)} className="btn btn-outline">
                                        <FaEdit style={{ marginRight: '0.5rem' }} /> Edit Profile
                                    </button>
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.5rem' }}>
                                    <InfoItem label="Date of Birth" value={new Date(profile.date_of_birth).toLocaleDateString()} />
                                    <InfoItem label="Gender" value={profile.gender} />
                                    <InfoItem label="Phone" value={profile.phone} />
                                    <InfoItem label="Address" value={profile.address} />
                                </div>
                            </div>

                            {/* Share Section */}
                            <div className="card" style={{ background: '#ECFDF5', borderColor: '#D1FAE5', marginBottom: '2rem' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                                    <div style={{ background: 'var(--white)', padding: '0.75rem', borderRadius: '50%' }}>
                                        <FaShareAlt size={24} color="var(--primary)" />
                                    </div>
                                    <div>
                                        <h3 style={{ fontSize: '1.25rem', fontWeight: 600, color: 'var(--primary-dark)' }}>Share Medical History</h3>
                                        <p style={{ fontSize: '0.875rem', color: 'var(--primary-dark)' }}>Grant temporary access to doctors via link or QR code.</p>
                                    </div>
                                </div>

                                {!shareData ? (
                                    <button onClick={generateShareLink} className="btn btn-primary">
                                        Generate Share Link
                                    </button>
                                ) : (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0 }}
                                        animate={{ opacity: 1, height: 'auto' }}
                                        style={{ background: 'var(--white)', padding: '1.5rem', borderRadius: 'var(--radius)', marginTop: '1rem' }}
                                    >
                                        <label className="input-label" style={{ fontSize: '0.875rem' }}>Public Share Link</label>
                                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                                            <input
                                                readOnly
                                                value={`${window.location.origin}/share/${shareData.share_token}`}
                                                className="input-field"
                                                style={{ background: '#F9FAFB' }}
                                            />
                                            <button
                                                className="btn btn-primary"
                                                onClick={() => navigator.clipboard.writeText(`${window.location.origin}/share/${shareData.share_token}`)}
                                            >
                                                Copy
                                            </button>
                                        </div>
                                        <div style={{ marginTop: '1rem', display: 'flex', justifyContent: 'center' }}>
                                            {/* Placeholder for QR Code - In a real app use a QR lib */}
                                            <div style={{ textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.875rem' }}>
                                                <FaQrcode size={120} color="var(--text-main)" />
                                                <p style={{ marginTop: '0.5rem' }}>Scan to view</p>
                                            </div>
                                        </div>
                                    </motion.div>
                                )}
                            </div>

                            {/* Timeline Section */}
                            <div style={{ marginBottom: '2rem' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                                    <h3 style={{ fontSize: '1.25rem', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <FaCalendarPlus color="var(--primary)" /> Health Timeline
                                    </h3>
                                    <button onClick={() => { setError(''); setShowAddVisit(true); }} className="btn btn-primary" style={{ fontSize: '0.875rem' }}>
                                        + Add Visit
                                    </button>
                                </div>

                                {visits.length === 0 ? (
                                    <div className="card" style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
                                        <FaHospital size={40} style={{ marginBottom: '1rem', opacity: 0.3 }} />
                                        <p>No hospital visits recorded yet.</p>
                                    </div>
                                ) : (
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                                        {visits.map((visit) => (
                                            <VisitCard
                                                key={visit.id}
                                                visit={visit}
                                                prescriptions={prescriptions.filter(p => p.visit_id === visit.id)}
                                                onDownload={downloadPrescription}
                                            />
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Add Visit Modal */}
                    <AnimatePresence>
                        {showAddVisit && (
                            <div className="modal-overlay" style={{
                                position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
                                background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center',
                                justifyContent: 'center', zIndex: 1000, padding: '1rem'
                            }}>
                                <motion.div
                                    initial={{ opacity: 0, scale: 0.95 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    exit={{ opacity: 0, scale: 0.95 }}
                                    className="card"
                                    style={{ maxWidth: '600px', width: '100%', maxHeight: '90vh', overflowY: 'auto' }}
                                >
                                    <h2 className="title" style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Add Hospital Visit</h2>
                                    {error && <div className="error-msg" style={{ marginBottom: '1.5rem' }}>{error}</div>}
                                    <form onSubmit={handleAddVisit}>
                                        <div className="input-group">
                                            <label className="input-label">Hospital Name</label>
                                            <input
                                                className="input-field"
                                                required
                                                value={newVisit.hospital_name}
                                                onChange={(e) => setNewVisit({ ...newVisit, hospital_name: e.target.value })}
                                                placeholder="e.g. City General Hospital"
                                            />
                                        </div>
                                        <div className="input-group">
                                            <label className="input-label">Visit Date</label>
                                            <input
                                                type="date"
                                                className="input-field"
                                                required
                                                value={newVisit.visit_date}
                                                onChange={(e) => setNewVisit({ ...newVisit, visit_date: e.target.value })}
                                            />
                                        </div>
                                        <div className="input-group">
                                            <label className="input-label">Reason for Visit</label>
                                            <input
                                                className="input-field"
                                                required
                                                value={newVisit.reason}
                                                onChange={(e) => setNewVisit({ ...newVisit, reason: e.target.value })}
                                                placeholder="e.g. Regular Checkup, Fever, etc."
                                            />
                                        </div>
                                        <div className="input-group">
                                            <label className="input-label">Notes (Optional)</label>
                                            <textarea
                                                className="input-field"
                                                rows="3"
                                                value={newVisit.notes}
                                                onChange={(e) => setNewVisit({ ...newVisit, notes: e.target.value })}
                                                placeholder="Any additional details..."
                                            />
                                        </div>
                                        <div className="input-group">
                                            <label className="input-label">Attach Prescription (Optional)</label>
                                            <input
                                                type="file"
                                                className="input-field"
                                                accept=".pdf,.jpg,.jpeg,.png"
                                                onChange={(e) => setVisitFile(e.target.files[0])}
                                            />
                                            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
                                                Accepted formats: PDF, JPG, PNG (Max 10MB)
                                            </p>
                                        </div>
                                        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end', marginTop: '1rem' }}>
                                            <button
                                                type="button"
                                                onClick={() => setShowAddVisit(false)}
                                                className="btn btn-outline"
                                                disabled={isSubmittingVisit}
                                            >
                                                Cancel
                                            </button>
                                            <button
                                                type="submit"
                                                className="btn btn-primary"
                                                disabled={isSubmittingVisit}
                                            >
                                                {isSubmittingVisit ? 'Adding...' : 'Add Visit'}
                                            </button>
                                        </div>
                                    </form>
                                </motion.div>
                            </div>
                        )}
                    </AnimatePresence>

                    {profile && isEditing && (
                        <div className="card" style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <h2 className="title" style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Edit Profile</h2>
                            <ProfileForm
                                formData={formData}
                                setFormData={setFormData}
                                handleSave={handleSave}
                                onCancel={() => setIsEditing(false)}
                            />
                        </div>
                    )}
                </motion.div>
            </main>
        </div>
    );
};

const VisitCard = ({ visit, prescriptions, onDownload }) => {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="card" style={{ padding: '0', overflow: 'hidden' }}>
            <div
                onClick={() => setExpanded(!expanded)}
                style={{
                    padding: '1.25rem', cursor: 'pointer', display: 'flex',
                    justifyContent: 'space-between', alignItems: 'center',
                    background: expanded ? 'rgba(16, 185, 129, 0.05)' : 'transparent'
                }}
            >
                <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                    <div style={{
                        background: 'var(--primary-light)', color: 'var(--primary)',
                        padding: '0.75rem', borderRadius: '0.75rem', display: 'flex', alignItems: 'center'
                    }}>
                        <FaHospital size={20} />
                    </div>
                    <div>
                        <h4 style={{ fontWeight: 600, fontSize: '1rem' }}>{visit.hospital_name}</h4>
                        <p style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                            {new Date(visit.visit_date).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })}
                        </p>
                    </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                    <span style={{
                        fontSize: '0.75rem', background: 'var(--bg-light)',
                        padding: '0.25rem 0.75rem', borderRadius: '1rem', fontWeight: 500
                    }}>
                        {visit.reason}
                    </span>
                    {expanded ? <FaChevronUp color="var(--text-muted)" /> : <FaChevronDown color="var(--text-muted)" />}
                </div>
            </div>

            <AnimatePresence>
                {expanded && (
                    <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        style={{ borderTop: '1px solid var(--border)', overflow: 'hidden' }}
                    >
                        <div style={{ padding: '1.25rem' }}>
                            {visit.notes && (
                                <div style={{ marginBottom: '1.5rem' }}>
                                    <h5 style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <FaNotesMedical /> Notes
                                    </h5>
                                    <p style={{ fontSize: '0.925rem', lineHeight: 1.5 }}>{visit.notes}</p>
                                </div>
                            )}

                            <div>
                                <h5 style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                    <FaFileMedical /> Prescriptions
                                </h5>
                                {prescriptions.length === 0 ? (
                                    <p style={{ fontSize: '0.875rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>No prescriptions attached.</p>
                                ) : (
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                        {prescriptions.map((p) => (
                                            <div key={p.id} style={{
                                                display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                                                background: 'var(--bg-light)', padding: '0.75rem 1rem', borderRadius: 'var(--radius)'
                                            }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                                    <FaFileMedical color="var(--primary)" />
                                                    <div>
                                                        <p style={{ fontSize: '0.875rem', fontWeight: 500 }}>{p.file_name}</p>
                                                        <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                                                            {(p.file_size / 1024).toFixed(1)} KB • {new Date(p.upload_date).toLocaleDateString()}
                                                        </p>
                                                    </div>
                                                </div>
                                                <button
                                                    onClick={(e) => { e.stopPropagation(); onDownload(p.id, p.file_name); }}
                                                    className="btn btn-outline"
                                                    style={{ padding: '0.4rem', borderRadius: '50%' }}
                                                    title="Download"
                                                >
                                                    <FaDownload size={14} />
                                                </button>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
};

const InfoItem = ({ label, value }) => (
    <div>
        <span style={{ display: 'block', fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>{label}</span>
        <span style={{ fontWeight: 500 }}>{value || 'Not set'}</span>
    </div>
);

const ProfileForm = ({ formData, setFormData, handleSave, isNew, onCancel }) => {
    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    return (
        <form onSubmit={handleSave}>
            <div className="input-group">
                <label className="input-label">Full Name</label>
                <input name="full_name" value={formData.full_name} onChange={handleChange} className="input-field" required />
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="input-group">
                    <label className="input-label">Date of Birth</label>
                    <input type="date" name="date_of_birth" value={formData.date_of_birth} onChange={handleChange} className="input-field" required />
                </div>
                <div className="input-group">
                    <label className="input-label">Gender</label>
                    <select name="gender" value={formData.gender} onChange={handleChange} className="input-field">
                        <option value="Male">Male</option>
                        <option value="Female">Female</option>
                        <option value="Other">Other</option>
                    </select>
                </div>
            </div>
            <div className="input-group">
                <label className="input-label">Phone</label>
                <input name="phone" value={formData.phone} onChange={handleChange} className="input-field" />
            </div>
            <div className="input-group">
                <label className="input-label">Address</label>
                <textarea name="address" value={formData.address} onChange={handleChange} className="input-field" rows="3" />
            </div>
            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
                {!isNew && (
                    <button type="button" onClick={onCancel} className="btn btn-outline">Cancel</button>
                )}
                <button type="submit" className="btn btn-primary">Save Profile</button>
            </div>
        </form>
    );
};

export default Profile;

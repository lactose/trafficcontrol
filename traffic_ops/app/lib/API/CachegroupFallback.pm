package API::CachegroupFallback;
#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#
#
# a note about locations and cachegroups. This used to be "Location", before we had physical locations in 12M. Very confusing.
# What used to be called a location is now called a "cache group" and location is now a physical address, not a group of caches working together.
#

# JvD Note: you always want to put Utils as the first use. Sh*t don't work if it's after the Mojo lines.
use UI::Utils;
use Mojo::Base 'Mojolicious::Controller';
use Data::Dumper;
use JSON;
use MojoPlugins::Response;
use Validate::Tiny ':all';

sub delete {
	my $self = shift;
	my $cache_id = $self->param('cacheGroupId');
	my $fallback_id = $self->param('fallbackId');
	my $params = $self->req->json;
	my $rs_backups = undef; 

	if ( !&is_oper($self) ) {
		return $self->forbidden();
	}

	if ( defined ($cache_id) && defined($fallback_id) ) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search( { primary_cg => $cache_id , backup_cg => $fallback_id} );
	} elsif (defined ($cache_id)) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search( { primary_cg => $cache_id} );
	} elsif (defined ($fallback_id)) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search( { backup_cg => $fallback_id} );
	}

	if ( ($rs_backups->count > 0) ) {
		my $del_records = $rs_backups->delete();
		if ($del_records) {
			&log( $self, "Backup configuration DELETED", "APICHANGE");
			return $self->success_message("Backup configuration DELETED");
		} else {
			return $self->alert( "Backup configuration DELETED." );
		}
	} else {
		&log( $self, "No backup Cachegroups found");
		return $self->not_found();
	}
}

sub show {
	my $self = shift;
	my $cache_id = $self->param("cacheGroupId");
	my $fallback_id = $self->param("fallbackId");
	my $id = $cache_id ? $cache_id : $fallback_id;

	#only integers
	if ( $id !~ /^\d+?$/ ) {
		&log( $self, "No such Cachegroup id $id");
		return $self->success([]);
	}

	my $cachegroup = $self->db->resultset('Cachegroup')->search( { id => $id } )->single();
	if ( !defined($cachegroup) ) {
		&log( $self, "No such Cachegroup $id");
		return $self->success([]);
	}

	if ( ($cachegroup->type->name ne "EDGE_LOC") ) {
		&log( $self, "cachegroup should be type EDGE_LOC.");
		return $self->success([]);
	}

	my $rs_backups = undef;

	if ( defined ($cache_id) && defined ($fallback_id)) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search({ primary_cg => $cache_id, backup_cg => $fallback_id}, {order_by => 'set_order'});
	} elsif ( defined ($cache_id) ) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search({ primary_cg => $cache_id}, {order_by => 'set_order'});
	} elsif ( defined ($fallback_id) ) {
		$rs_backups = $self->db->resultset('CachegroupFallback')->search({ backup_cg => $fallback_id}, {order_by => 'set_order'});
	}

	if ( defined ($rs_backups) && ($rs_backups->count > 0) ) {
		my $response;
		my $backup_cnt = 0;
		while ( my $row = $rs_backups->next ) {
			$response->[$backup_cnt]{"cacheGroupId"} = $row->primary_cg->id;
			$response->[$backup_cnt]{"cacheGroupName"} = $row->primary_cg->name;
			$response->[$backup_cnt]{"fallbackName"} = $row->backup_cg->name;
			$response->[$backup_cnt]{"fallbackId"} = $row->backup_cg->id;
			$response->[$backup_cnt]{"fallbackOrder"} = $row->set_order;
			$backup_cnt++;
		}
		return $self->success( $response );
	} else {
		&log( $self, "No backup Cachegroups");
		return $self->success([]);
	}
}

sub create {
	my $self = shift;
	my $cache_id = $self->param('cacheGroupId');
	my $params = $self->req->json;

	if ( !defined($cache_id)) {
		my @param_array = @{$params};
		$cache_id = $param_array[0]{cacheGroupId};
	}

	if ( !defined($params) ) {
		return $self->alert("parameters must be in JSON format,  please check!");
	}

	if ( !&is_oper($self) ) {
		return $self->forbidden();
	}

	#only integers
	if ( $cache_id !~ /^\d+?$/ ) {
		&log( $self, "No such Cachegroup id $cache_id");
		return $self->not_found();
	}

	my $cachegroup = $self->db->resultset('Cachegroup')->search( { id => $cache_id } )->single();
	if ( !defined($cachegroup) ) {
		return $self->not_found();
	}

	if ( ($cachegroup->type->name ne "EDGE_LOC") ) {
		return $self->alert("cachegroup should be type EDGE_LOC.");
	}

	foreach my $config (@{ $params }) {
		my $rs_backup = $self->db->resultset('Cachegroup')->search( { id => $config->{fallbackId} } )->single();
		if ( !defined($rs_backup) ) {
			&log( $self, "ERROR Backup config: No such Cache Group $config->{fallbackId}");
			next;
		} 

		if ( ($rs_backup->type->name ne "EDGE_LOC") ) {
			&log( $self, "ERROR Backup config: $config->{name} is not EDGE_LOC");
			next;
		}

		my $existing_row = $self->db->resultset('CachegroupFallback')->search( { primary_cg => $cache_id, backup_cg => $config->{fallbackId} } );
		if ( defined ($existing_row->next) ) {
			next;#Skip existing rows
		}

		my $values = {
			primary_cg => $cache_id ,
			backup_cg  => $config->{fallbackId},
			set_order  => $config->{fallbackOrder}
		};
        
		my $rs_data = $self->db->resultset('CachegroupFallback')->create($values)->insert();
		if ( !defined($rs_data)) {
			&log( $self, "Database operation for backup configuration for cache group $cache_id failed.");
		}
	}

	my $rs_backups = $self->db->resultset('CachegroupFallback')->search({ primary_cg => $cache_id}, {order_by => 'set_order'});
	my $response;
	my $backup_cnt = 0;
	if ( ($rs_backups->count > 0) ) {
		while ( my $row = $rs_backups->next ) {
			$response->[$backup_cnt]{"cacheGroupId"}   = $cache_id;
			$response->[$backup_cnt]{"cacheGroupName"} = $row->primary_cg->name;
			$response->[$backup_cnt]{"fallbackName"}   = $row->backup_cg->name;
			$response->[$backup_cnt]{"fallbackId"}     = $row->backup_cg->id;
			$response->[$backup_cnt]{"fallbackOrder"}  = $row->set_order;
			$backup_cnt++;
		}
		&log( $self, "Backup configuration UPDATED for $cache_id", "APICHANGE");
		return $self->success( $response, "Backup configuration CREATE for cache group $cache_id successful." );
	} else {
		return $self->alert("Backup configuration CREATE for cache group $cache_id Failed." );
	}
}


sub update {
	my $self = shift;
	my $cache_id = $self->param('cacheGroupId');
	my $params = $self->req->json;

	if ( !defined($cache_id)) {
		my @param_array = @{$params};
		$cache_id = $param_array[0]{cacheGroupId};
	}

	if ( !defined($params) ) {
		return $self->alert("parameters must be in JSON format,  please check!");
	}

	if ( !&is_oper($self) ) {
		return $self->forbidden();
	}

	#only integers
	if ( $cache_id !~ /^\d+?$/ ) {
		&log( $self, "No such Cachegroup id $cache_id");
		return $self->not_found();
	}

	my $cachegroup = $self->db->resultset('Cachegroup')->search( { id => $cache_id } )->single();
	if ( !defined($cachegroup) ) {
		return $self->not_found();
	}

	if ( ($cachegroup->type->name ne "EDGE_LOC") ) {
		return $self->alert("cachegroup should be type EDGE_LOC.");
	}

	my $rs_backups = $self->db->resultset('CachegroupFallback')->search( { primary_cg => $cache_id } );
	if ( !defined ($rs_backups->next) ) {
		return $self->alert( "Backup list not configured for $cache_id, create and update" );
	}

	foreach my $config (@{ $params }) {
		my $rs_backup = $self->db->resultset('Cachegroup')->search( { id => $config->{fallbackId} } )->single();
		if ( !defined($rs_backup) ) {
			&log( $self, "ERROR Backup config: No such Cache Group $config->{fallbackId}");
			next;
		} 

		if ( ($rs_backup->type->name ne "EDGE_LOC") ) {
			&log( $self, "ERROR Backup config: $config->{name} is not EDGE_LOC");
			next;
		}

		my $values = {
			primary_cg => $cache_id ,
			backup_cg  => $config->{fallbackId},
			set_order  => $config->{fallbackOrder}
		};

		my $existing_row = $self->db->resultset('CachegroupFallback')->search( { primary_cg => $cache_id, backup_cg => $config->{fallbackId} } );
		#New row creation disabled for PUT.Only existing rows can be updated
		if ( defined ($existing_row->next) ) {
			$existing_row->update($values);
		}
	}

	my $rs_backups = $self->db->resultset('CachegroupFallback')->search({ primary_cg => $cache_id}, {order_by => 'set_order'});
	my $response;
	my $backup_cnt = 0;
	if ( ($rs_backups->count > 0) ) {
		while ( my $row = $rs_backups->next ) {
			$response->[$backup_cnt]{"cacheGroupId"}   = $cache_id;
			$response->[$backup_cnt]{"cacheGroupName"} = $row->primary_cg->name;
			$response->[$backup_cnt]{"fallbackName"}   = $row->backup_cg->name;
			$response->[$backup_cnt]{"fallbackId"}     = $row->backup_cg->id;
			$response->[$backup_cnt]{"fallbackOrder"}  = $row->set_order;
			$backup_cnt++;
		}
		&log( $self, "Backup configuration UPDATED for $cache_id", "APICHANGE");
		return $self->success( $response, "Backup configuration UPDATE for cache group $cache_id successful." );
	} else {
		return $self->alert("Backup configuration UPDATE for cache group $cache_id Failed." );
	}
}

1;